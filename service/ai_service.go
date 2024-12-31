package service

import (
	"a21hc3NpZ25tZW50/model"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type AIService struct {
	Client HTTPClient
}

func (s *AIService) AnalyzeData(table map[string][]string, query, token string) (string, error) {
	url := "https://api-inference.huggingface.co/models/google/tapas-base-finetuned-wtq"

	if len(table) == 0 {
		return "", errors.New("table cannot be empty")
	}
	// fmt.Println(table)

	requestBody := model.AIRequest{
		Inputs: model.Inputs{
			Table: table,
			Query: query,
		},
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	resp, err := s.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed with status: %s", resp.StatusCode)
	}
	// bodyBytes, _ := io.ReadAll(resp.Body)

	var resTapas model.TapasResponse
	err = json.NewDecoder(resp.Body).Decode(&resTapas)
	if err != nil {
		return "", err
	}

	if len(resTapas.Answer) == 0 {
		return "", errors.New("no answer received from AI service")
	}

	// result := string(bodyBytes)
	fmt.Println(resTapas.Coordinates)
	fmt.Println(resTapas.Cells)
	return resTapas.Answer, nil
	// return result, nil
}

func (s *AIService) ChatWithAI(context, query, token string) (model.ChatResponse, error) {
	url := "https://api-inference.huggingface.co/models/microsoft/Phi-3.5-mini-instruct/v1/chat/completions"

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type Choice struct {
		Index        int         `json:"index"`
		Message      Message     `json:"message"`
		Logprobs     interface{} `json:"logprobs"` // Use interface{} if the type is unknown or nil
		FinishReason string      `json:"finish_reason"`
	}

	type Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	}

	type Response struct {
		Object            string   `json:"object"`
		ID                string   `json:"id"`
		Created           int64    `json:"created"`
		Model             string   `json:"model"`
		SystemFingerprint string   `json:"system_fingerprint"`
		Choices           []Choice `json:"choices"`
		Usage             Usage    `json:"usage"`
	}

	payload := map[string]any{
		"model": "microsoft/Phi-3.5-mini-instruct",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": query,
			},
		},
		"max_tokens": 500,
		"stream":     false,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return model.ChatResponse{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return model.ChatResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send the request
	resp, err := s.Client.Do(req)
	if err != nil {
		return model.ChatResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.ChatResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error respons
	if resp.StatusCode != http.StatusOK {
		return model.ChatResponse{}, fmt.Errorf("failed with status: %s, response: %s", resp.Status, string(body))
	}

	// body value check
	if len(body) == 0 {
		return model.ChatResponse{}, errors.New("received empty response body")
	}

	// var chatResponse []model.ChatResponse
	// if err := json.Unmarshal(body, &chatResponse); err != nil {
	// 	return model.ChatResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	// }

	// if len(chatResponse) == 0 || strings.TrimSpace(chatResponse[0].GeneratedText) == "" {
	// 	return model.ChatResponse{}, errors.New("received empty or invalid response from Chat API")
	// }

	// chatResponse[0].GeneratedText = strings.TrimSpace(chatResponse[0].GeneratedText)
	// return chatResponse[0], nil

	var result model.ChatResponse
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return model.ChatResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	result.GeneratedText = response.Choices[0].Message.Content

	return result, nil
}
