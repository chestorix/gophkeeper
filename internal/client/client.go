// internal/client/client.go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/chestorix/gophkeeper/internal/models"
)

type Client struct {
	config     *Config
	httpClient *http.Client
	token      string
}

func NewClient(cfg *Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: cfg.Token,
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reader = bytes.NewReader(jsonData)
	}

	url := "http://" + c.config.ServerAddress + path
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.httpClient.Do(req)
}

func (c *Client) Register(ctx context.Context, login, password string) (*models.AuthResponse, error) {
	req := models.AuthRequest{
		Login:    login,
		Password: password,
	}

	resp, err := c.makeRequest(ctx, "POST", "/api/user/register", req)
	if err != nil {
		return nil, fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registration failed: %s", string(body))
	}

	var authResp models.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.token = authResp.Token
	return &authResp, nil
}

func (c *Client) Login(ctx context.Context, login, password string) (*models.AuthResponse, error) {
	req := models.AuthRequest{
		Login:    login,
		Password: password,
	}

	resp, err := c.makeRequest(ctx, "POST", "/api/user/login", req)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed: %s", string(body))
	}

	var authResp models.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.token = authResp.Token
	return &authResp, nil
}

func (c *Client) SaveData(ctx context.Context, data *models.SecretItemData) error {
	resp, err := c.makeRequest(ctx, "POST", "/api/data", data)
	if err != nil {
		return fmt.Errorf("save data request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("save data failed: %s", string(body))
	}

	return nil
}

func (c *Client) GetData(ctx context.Context, dataID string) (*models.SecretItemData, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/data/"+dataID, nil)
	if err != nil {
		return nil, fmt.Errorf("get data request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get data failed: %s", string(body))
	}

	var data models.SecretItemData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &data, nil
}

func (c *Client) ListData(ctx context.Context, lastSync time.Time) ([]models.SecretItemData, error) {
	path := "/api/data"
	if !lastSync.IsZero() {
		path += "?last_sync=" + lastSync.Format(time.RFC3339)
	}

	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("list data request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list data failed: %s", string(body))
	}

	var data []models.SecretItemData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return data, nil
}

func (c *Client) UpdateData(ctx context.Context, data *models.SecretItemData) error {
	resp, err := c.makeRequest(ctx, "PUT", "/api/data/"+data.ID, data)
	if err != nil {
		return fmt.Errorf("update data request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update data failed: %s", string(body))
	}

	return nil
}

func (c *Client) DeleteData(ctx context.Context, dataID string) error {
	resp, err := c.makeRequest(ctx, "DELETE", "/api/data/"+dataID, nil)
	if err != nil {
		return fmt.Errorf("delete data request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete data failed: %s", string(body))
	}

	return nil
}

func (c *Client) SyncData(ctx context.Context, req *models.SyncRequest) (*models.SyncResponse, error) {
	resp, err := c.makeRequest(ctx, "POST", "/api/sync", req)
	if err != nil {
		return nil, fmt.Errorf("sync data request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sync data failed: %s", string(body))
	}

	var syncResp models.SyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &syncResp, nil
}

func (c *Client) Health(ctx context.Context) error {
	resp, err := c.makeRequest(ctx, "GET", "/health", nil)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server health check failed: status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetDataByName(ctx context.Context, name string) (*models.SecretItemData, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/data/name/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("get data by name request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get data by name failed: %s", string(body))
	}

	var data models.SecretItemData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &data, nil
}

func (c *Client) DeleteDataByName(ctx context.Context, name string) error {
	resp, err := c.makeRequest(ctx, "DELETE", "/api/data/name/"+name, nil)
	if err != nil {
		return fmt.Errorf("delete data by name request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete data by name failed: %s", string(body))
	}

	return nil
}
