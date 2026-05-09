package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client calls users-service HTTP APIs used by feed-service.
type Client struct {
	base   string
	client *http.Client
}

func NewClient(baseURL string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return &Client{
		base: baseURL,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type successEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

type followersBody struct {
	FollowerIDs []string `json:"follower_ids"`
}

// FollowerIDs returns ids of users who follow authorID (followee / post author).
func (c *Client) FollowerIDs(ctx context.Context, authorID string) ([]string, error) {
	if c.base == "" {
		return nil, fmt.Errorf("users client: empty base URL")
	}
	u := fmt.Sprintf("%s/api/v1/users/%s/followers", c.base, url.PathEscape(authorID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("users service: GET followers status %d", resp.StatusCode)
	}
	var env successEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	if !env.Success {
		return nil, fmt.Errorf("users service: success=false on followers response")
	}
	var body followersBody
	if err := json.Unmarshal(env.Data, &body); err != nil {
		return nil, fmt.Errorf("decode followers data: %w", err)
	}
	if body.FollowerIDs == nil {
		return []string{}, nil
	}
	return body.FollowerIDs, nil
}
