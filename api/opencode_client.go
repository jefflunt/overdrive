package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type OpenCodeSession struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type OpenCodeMessage struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type OpenCodeClient struct {
	BaseURL string
	HTTP    *http.Client
}

func NewOpenCodeClient(port string) *OpenCodeClient {
	return NewOpenCodeClientWithTimeout(port, 10*time.Second)
}

func NewOpenCodeClientWithTimeout(port string, timeout time.Duration) *OpenCodeClient {
	return &OpenCodeClient{
		BaseURL: fmt.Sprintf("http://127.0.0.1:%s", strings.TrimSpace(port)),
		HTTP: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *OpenCodeClient) getClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

func (c *OpenCodeClient) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/global/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.getClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenCodeClient) CreateSession(title string) (string, error) {
	payload := map[string]string{"title": title}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := c.getClient().Post(c.BaseURL+"/session", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("create session failed with status: %d", resp.StatusCode)
	}

	var session OpenCodeSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return "", err
	}
	return session.ID, nil
}

func (c *OpenCodeClient) SendMessageAsync(sessionID, content string) error {
	log.Printf("[DEBUG] OpenCodeClient.SendMessageAsync: sessionID=%s, content=%s", sessionID, content)
	payload := map[string]interface{}{
		"parts": []map[string]string{
			{"type": "text", "text": content},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.getClient().Post(c.BaseURL+"/session/"+sessionID+"/prompt_async", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[ERROR] OpenCodeClient.SendMessageAsync: POST failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] OpenCodeClient.SendMessageAsync: status=%d", resp.StatusCode)
		return fmt.Errorf("send message async failed with status: %d", resp.StatusCode)
	}
	log.Printf("[DEBUG] OpenCodeClient.SendMessageAsync: Success")
	return nil
}

func (c *OpenCodeClient) IsGenerating(sessionID string) (bool, error) {
	resp, err := c.getClient().Get(c.BaseURL + "/session/status")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("get status failed: %d", resp.StatusCode)
	}

	var statuses map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&statuses); err != nil {
		return false, err
	}

	status, ok := statuses[sessionID]
	if !ok {
		log.Printf("[DEBUG] OpenCodeClient.IsGenerating: session %s not found in status", sessionID)
		return false, nil
	}

	s := fmt.Sprintf("%v", status)
	log.Printf("[DEBUG] OpenCodeClient.IsGenerating: session %s status=%s", sessionID, s)
	return s != "idle" && s != "ready", nil
}

func (c *OpenCodeClient) GetSessionMessages(sessionID string) ([]OpenCodeMessage, error) {
	log.Printf("[DEBUG] OpenCodeClient.GetSessionMessages: sessionID=%s", sessionID)
	type Part struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
	}
	type MessageResponse struct {
		Info struct {
			ID        string    `json:"id"`
			Role      string    `json:"role"`
			Timestamp time.Time `json:"timestamp"`
		} `json:"info"`
		Parts []Part `json:"parts"`
	}

	resp, err := c.getClient().Get(c.BaseURL + "/session/" + sessionID + "/message")
	if err != nil {
		log.Printf("[ERROR] OpenCodeClient.GetSessionMessages: GET failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] OpenCodeClient.GetSessionMessages: status=%d", resp.StatusCode)
		return nil, fmt.Errorf("get messages failed with status: %d", resp.StatusCode)
	}

	var responses []MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		log.Printf("[ERROR] OpenCodeClient.GetSessionMessages: decode failed: %v", err)
		return nil, err
	}

	var messages []OpenCodeMessage
	for _, r := range responses {
		content := ""
		for _, p := range r.Parts {
			if p.Type == "text" {
				content += p.Text
			}
		}
		messages = append(messages, OpenCodeMessage{
			ID:        r.Info.ID,
			SessionID: sessionID,
			Role:      r.Info.Role,
			Content:   content,
			Timestamp: r.Info.Timestamp,
		})
	}
	log.Printf("[DEBUG] OpenCodeClient.GetSessionMessages: returning %d messages", len(messages))
	return messages, nil
}

func (c *OpenCodeClient) WaitForHealth(timeout time.Duration) error {
	start := time.Now()
	for {
		if err := c.Health(); err == nil {
			return nil
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for health")
		}
		time.Sleep(500 * time.Millisecond)
	}
}
