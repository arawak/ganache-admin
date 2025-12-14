package ganache

import (
	"encoding/json"
	"fmt"
	"time"
)

type Asset struct {
	ID         StringID  `json:"id"`
	Title      string    `json:"title"`
	Caption    string    `json:"caption"`
	Credit     string    `json:"credit"`
	Source     string    `json:"source"`
	UsageNotes string    `json:"usageNotes"`
	Tags       []string  `json:"tags"`
	Variants   Variants  `json:"variants"`
	CreatedAt  time.Time `json:"createdAt"`
}

type StringID string

func (s *StringID) UnmarshalJSON(data []byte) error {
	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		*s = StringID(asString)
		return nil
	}
	var asNumber json.Number
	if err := json.Unmarshal(data, &asNumber); err == nil {
		*s = StringID(asNumber.String())
		return nil
	}
	return fmt.Errorf("invalid id: %s", string(data))
}

type Variants struct {
	Thumb    string `json:"thumb"`
	Content  string `json:"content"`
	Original string `json:"original"`
}

type AssetUpdate struct {
	Title      string   `json:"title"`
	Caption    string   `json:"caption"`
	Credit     string   `json:"credit"`
	Source     string   `json:"source"`
	UsageNotes string   `json:"usageNotes"`
	Tags       []string `json:"tags"`
}

type SearchResponse struct {
	Assets   []Asset `json:"assets"`
	Items    []Asset `json:"items"`
	Page     int     `json:"page"`
	PageSize int     `json:"pageSize"`
	Total    int     `json:"total"`
}

type ErrorResponse struct {
	Error   ErrorBody `json:"error"`
	Message string    `json:"message"`
}

type ErrorBody struct {
	Message string `json:"message"`
}

type Tag struct {
	Name string `json:"name"`
}

type TagResponse struct {
	Tags []Tag `json:"tags"`
}
