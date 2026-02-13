package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Client struct {
	api v1.API
}

func NewClient(address string) (*Client, error) {
	if address == "" {
		return nil, fmt.Errorf("prometheus address is empty")
	}
	cfg := api.Config{Address: address}
	cli, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create prometheus client: %w", err)
	}
	return &Client{api: v1.NewAPI(cli)}, nil
}

func (c *Client) Query(ctx context.Context, query string, ts time.Time) (model.Value, v1.Warnings, error) {
	if c == nil || c.api == nil {
		return nil, nil, fmt.Errorf("prometheus client not configured")
	}
	return c.api.Query(ctx, query, ts)
}
