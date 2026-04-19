package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with helper methods.
type Client struct {
	api *vaultapi.Client
}

// NewClient creates a new Vault client using the provided address and token.
func NewClient(address, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	api.SetToken(token)

	return &Client{api: api}, nil
}

// GetSecrets reads KV v2 secrets at the given path and returns them as a
// map of key → value strings.
func (c *Client) GetSecrets(ctx context.Context, secretPath string) (map[string]string, error) {
	// Normalise path: insert /data/ for KV v2 mounts (mount/data/path).
	mountPath, subPath, err := splitMountAndPath(secretPath)
	if err != nil {
		return nil, err
	}

	kvPath := fmt.Sprintf("%s/data/%s", mountPath, subPath)

	secret, err := c.api.KVv2(mountPath).Get(ctx, subPath)
	if err != nil {
		// Fall back to raw logical read for non-KV-v2 paths.
		logical, lerr := c.api.Logical().ReadWithContext(ctx, kvPath)
		if lerr != nil {
			return nil, fmt.Errorf("reading secret at %q: %w", secretPath, err)
		}
		if logical == nil || logical.Data == nil {
			return nil, fmt.Errorf("no secret found at path %q", secretPath)
		}
		return flattenData(logical.Data), nil
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no secret found at path %q", secretPath)
	}

	return flattenData(secret.Data), nil
}

// splitMountAndPath splits "mount/sub/path" into ("mount", "sub/path").
func splitMountAndPath(p string) (string, string, error) {
	parts := strings.SplitN(strings.Trim(p, "/"), "/", 2)
	if len(parts) < 2 || parts[1] == "" {
		return "", "", fmt.Errorf("secret path %q must be at least mount/key", p)
	}
	return parts[0], parts[1], nil
}

// flattenData converts map[string]interface{} values to strings.
func flattenData(data map[string]interface{}) map[string]string {
	out := make(map[string]string, len(data))
	for k, v := range data {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}
