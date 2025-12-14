package auth

import (
	"errors"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

type User struct {
	Username     string `yaml:"username"`
	PasswordHash string `yaml:"passwordHash"`
}

type UsersFile struct {
	Users []User `yaml:"users"`
}

type UserStore struct {
	users map[string]User
}

func LoadUsers(path string) (*UserStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var uf UsersFile
	if err := yaml.Unmarshal(data, &uf); err != nil {
		return nil, err
	}
	return NewUserStore(uf.Users)
}

func NewUserStore(list []User) (*UserStore, error) {
	users := make(map[string]User, len(list))
	for _, u := range list {
		if u.Username == "" || u.PasswordHash == "" {
			return nil, errors.New("username and passwordHash required")
		}
		users[u.Username] = u
	}
	return &UserStore{users: users}, nil
}

func (s *UserStore) Validate(username, password string) bool {
	user, ok := s.users[username]
	if !ok {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}
