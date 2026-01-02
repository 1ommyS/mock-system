package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mock-svc/internal/application"
)

type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *AuthClient) CheckAccess(ctx context.Context, token, resourceType, resourceID, action string) (bool, string, bool, error) {
	payload := map[string]string{
		"resourceType": resourceType,
		"resourceId":   resourceID,
		"action":       action,
	}
	resp, status, err := c.doJSON(ctx, http.MethodPost, "/authz/v1/check", token, payload)
	if err != nil {
		return false, "", false, err
	}
	defer resp.Body.Close()
	if status < 200 || status >= 300 {
		return false, "", false, fmt.Errorf("auth check failed: %s", resp.Status)
	}
	var out struct {
		Allowed             bool   `json:"allowed"`
		EffectivePermission string `json:"effectivePermission"`
		IsOwner             bool   `json:"isOwner"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, "", false, err
	}
	return out.Allowed, out.EffectivePermission, out.IsOwner, nil
}

func (c *AuthClient) ListResources(ctx context.Context, token, resourceType, minPermission string) ([]application.ResourceAccess, error) {
	q := url.Values{}
	q.Set("resourceType", resourceType)
	if minPermission != "" {
		q.Set("minPermission", minPermission)
	}
	path := "/authz/v1/resources?" + q.Encode()
	resp, status, err := c.doJSON(ctx, http.MethodGet, path, token, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("auth list resources failed: %s", resp.Status)
	}
	var out struct {
		Items []struct {
			ResourceID          string `json:"resourceId"`
			EffectivePermission string `json:"effectivePermission"`
			IsOwner             bool   `json:"isOwner"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	items := make([]application.ResourceAccess, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, application.ResourceAccess{
			ResourceID:          item.ResourceID,
			EffectivePermission: item.EffectivePermission,
			IsOwner:             item.IsOwner,
		})
	}
	return items, nil
}

func (c *AuthClient) RegisterResource(ctx context.Context, resourceType, resourceID, ownerUserID string) error {
	payload := map[string]string{
		"resourceType": resourceType,
		"resourceId":   resourceID,
		"ownerUserId":  ownerUserID,
	}
	resp, status, err := c.doJSON(ctx, http.MethodPost, "/authz/v1/resources", "", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if status < 200 || status >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth register failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *AuthClient) doJSON(ctx context.Context, method, path, token string, payload any) (*http.Response, int, error) {
	endpoint := c.baseURL + path
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, 0, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	return resp, resp.StatusCode, nil
}
