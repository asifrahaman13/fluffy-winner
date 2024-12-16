package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/asifrahaman13/bhagabad_gita/internal/core/domain"
	"github.com/asifrahaman13/bhagabad_gita/internal/core/ports"
	"github.com/asifrahaman13/bhagabad_gita/internal/helper"
	"github.com/asifrahaman13/bhagabad_gita/internal/config"
	"io"
	"net/http"
	"errors"
)

type userService struct {
	repo ports.UserRepository
}

func InitializeUserService(r ports.UserRepository) *userService {
	return &userService{
		repo: r,
	}
}

func (s *userService) Signup(user domain.User) (string, error) {
	message, err := s.repo.Create(user, "users")
	if err != nil {
		panic(err)
	}
	if !message {
		return "Failed to insert data", nil
	}
	return "Successfully stored the information", nil
}

func (s *userService) Login(user domain.User) (domain.AccessToken, error) {
	token, err := helper.CreateToken(user.Username, "user")
	if err != nil {
		panic(err)
	}
	accessToken := domain.AccessToken{
		Token: token,
	}
	return accessToken, nil
}

func (s *userService) GetLLMResponse(query string) (string, error) {
	config, err := config.NewConfig() 
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}
	posturl := config.LLamaUrl
	payload := map[string]interface{}{
		"model":  "llama3.1",
		"prompt": query,
		"stream": false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", posturl, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 HTTP response: %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	responseText := string(respBody)
	if responseText == "" {
		return "", errors.New("received empty response")
	}
	var data domain.ResponseData
	err = json.Unmarshal([]byte(responseText), &data)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return data.Response, nil
}