package ganache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) SearchAssets(ctx context.Context, q string, tags []string, page, pageSize int, sort string) (SearchResponse, error) {
	u, _ := url.Parse(c.baseURL)
	u.Path = path.Join(u.Path, "/api/assets")
	query := u.Query()
	if q != "" {
		query.Set("q", q)
	}
	for _, t := range tags {
		if t != "" {
			query.Add("tag", t)
		}
	}
	if page > 0 {
		query.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	}
	if sort != "" {
		query.Set("sort", sort)
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return SearchResponse{}, err
	}
	c.addAuth(req)

	var respData SearchResponse
	if err := c.doJSON(req, &respData); err != nil {
		return SearchResponse{}, err
	}
	if len(respData.Assets) == 0 && len(respData.Items) > 0 {
		respData.Assets = respData.Items
	}
	for i := range respData.Assets {
		c.absolutizeVariants(&respData.Assets[i])
	}
	return respData, nil
}

func (c *Client) GetAsset(ctx context.Context, id string) (Asset, error) {
	u := fmt.Sprintf("%s/api/assets/%s", c.baseURL, url.PathEscape(id))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Asset{}, err
	}
	c.addAuth(req)
	var asset Asset
	if err := c.doJSON(req, &asset); err != nil {
		return Asset{}, err
	}
	c.absolutizeVariants(&asset)
	return asset, nil
}

func (c *Client) UpdateAsset(ctx context.Context, id string, update AssetUpdate) (Asset, error) {
	body, err := json.Marshal(update)
	if err != nil {
		return Asset{}, err
	}
	u := fmt.Sprintf("%s/api/assets/%s", c.baseURL, url.PathEscape(id))
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u, bytes.NewReader(body))
	if err != nil {
		return Asset{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.addAuth(req)
	var asset Asset
	if err := c.doJSON(req, &asset); err != nil {
		return Asset{}, err
	}
	return asset, nil
}

func (c *Client) DeleteAsset(ctx context.Context, id string) error {
	u := fmt.Sprintf("%s/api/assets/%s", c.baseURL, url.PathEscape(id))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	c.addAuth(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return parseError(resp)
	}
	return nil
}

func (c *Client) CreateAssetMultipart(ctx context.Context, file io.Reader, filename string, fields map[string]string, tags []string) (Asset, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	fw, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return Asset{}, err
	}
	if _, err := io.Copy(fw, file); err != nil {
		return Asset{}, err
	}
	for k, v := range fields {
		if v == "" {
			continue
		}
		_ = writer.WriteField(k, v)
	}
	for _, t := range tags {
		if t == "" {
			continue
		}
		_ = writer.WriteField("tags[]", t)
	}
	writer.Close()

	u := fmt.Sprintf("%s/api/assets", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, &buf)
	if err != nil {
		return Asset{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.addAuth(req)
	var asset Asset
	if err := c.doJSON(req, &asset); err != nil {
		return Asset{}, err
	}
	return asset, nil
}

func (c *Client) ListTags(ctx context.Context, prefix string, page, pageSize int) (TagResponse, error) {
	u, _ := url.Parse(c.baseURL)
	u.Path = path.Join(u.Path, "/api/tags")
	q := u.Query()
	if prefix != "" {
		q.Set("prefix", prefix)
	}
	if page > 0 {
		q.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		q.Set("pageSize", fmt.Sprintf("%d", pageSize))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return TagResponse{}, err
	}
	c.addAuth(req)
	var respData TagResponse
	if err := c.doJSON(req, &respData); err != nil {
		return TagResponse{}, err
	}
	return respData, nil
}

func (c *Client) doJSON(req *http.Request, target any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return parseError(resp)
	}
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(target)
}

func (c *Client) addAuth(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("X-Api-Key", c.apiKey)
	}
}

func (c *Client) absolutizeVariants(a *Asset) {
	if a == nil {
		return
	}
	if strings.HasPrefix(a.Variants.Thumb, "/") {
		a.Variants.Thumb = c.baseURL + a.Variants.Thumb
	}
	if strings.HasPrefix(a.Variants.Content, "/") {
		a.Variants.Content = c.baseURL + a.Variants.Content
	}
	if strings.HasPrefix(a.Variants.Original, "/") {
		a.Variants.Original = c.baseURL + a.Variants.Original
	}
}

func parseError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)
	var er ErrorResponse
	if err := json.Unmarshal(data, &er); err == nil {
		if er.Error.Message != "" {
			return fmt.Errorf("%s", er.Error.Message)
		}
		if er.Message != "" {
			return fmt.Errorf("%s", er.Message)
		}
	}
	if len(data) > 0 {
		return fmt.Errorf("%s", strings.TrimSpace(string(data)))
	}
	return fmt.Errorf("unexpected status: %d", resp.StatusCode)
}
