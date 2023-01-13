package authstore

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type AuthStore[T any] interface {
	Load() (*T, error)
	Save(authData *T) error
	Exists() bool
}

type authStore[T any] struct {
	jsonFile string
}

func NewAuthStore[T any](jsonFile string) AuthStore[T] {
	return &authStore[T]{jsonFile: jsonFile}
}

func (s *authStore[T]) Load() (*T, error) {
	fileContent, err := ioutil.ReadFile(s.jsonFile)
	if err != nil {
		return nil, err
	}

	authData := new(T)
	err = json.Unmarshal(fileContent, authData)
	if err != nil {
		return nil, err
	}

	return authData, nil
}

func (s *authStore[T]) Save(authData *T) error {
	jsonContent, err := json.Marshal(authData)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(s.jsonFile), 0700)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(s.jsonFile, jsonContent, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (s *authStore[T]) Exists() bool {
	if _, err := os.Stat(s.jsonFile); os.IsNotExist(err) {
		return false
	}

	return true
}
