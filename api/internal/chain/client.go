package chain

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "time"
)

// Interface matches http handlers expectations.
type Interface interface {
    PushPrediction(ctx context.Context, astro, grav uint32) (string, error)
    GetComposite(ctx context.Context) (uint32, error)
    UpdateWeights(ctx context.Context, a, g, ml uint32) (string, error)
}

// Client implements Interface; mock vs real behavior depends on config.
type Client struct {
    cfg        Config
    httpClient *http.Client
}

func New(address, rpc, key string) *Client {
    cfg := Config{ContractAddress: address, RPCURL: rpc, PrivateKey: key, AccountAddress: "", Timeout: 3 * time.Second}
    return NewWithConfig(cfg)
}

func NewWithConfig(cfg Config) *Client {
    return &Client{cfg: cfg, httpClient: &http.Client{Timeout: cfg.Timeout}}
}

// PushPrediction submits prediction inputs (mocked or stubbed).
func (c *Client) PushPrediction(ctx context.Context, astro, grav uint32) (string, error) {
    if astro > 100 || grav > 100 {
        return "", errors.New("invalid score >100")
    }
    if !c.cfg.IsEnabled() {
        select {
        case <-time.After(15 * time.Millisecond):
            return "0xMOCK_PRED_TX", nil
        case <-ctx.Done():
            return "", ctx.Err()
        }
    }
    if c.cfg.RPCURL == "" || c.cfg.ContractAddress == "" {
        return "", errors.New("incomplete config for real tx")
    }
    // Minimal raw invoke_function payload (placeholder calldata; real encoding TBD)
    payload := map[string]any{
        "jsonrpc": "2.0",
        "method":  "starknet_addInvokeTransaction",
        "params": []any{map[string]any{
            "type": "INVOKE",
            "sender_address": c.cfg.AccountAddress,
            "calldata": []string{c.cfg.ContractAddress, "set_prediction_inputs", fmt.Sprintf("%d", astro), fmt.Sprintf("%d", grav)},
            "signature": []string{}, // TODO: sign transaction
            "max_fee": "0x0",
            "version": "0x1",
        }},
        "id": 1,
    }
    buf, _ := json.Marshal(payload)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.RPCURL, bytes.NewReader(buf))
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.httpClient.Do(req)
    if err != nil { return "", err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return "", errors.New("rpc_unavailable") }
    var out struct{ Result any `json:"result"`; Error any `json:"error"` }
    _ = json.NewDecoder(resp.Body).Decode(&out)
    if out.Error != nil { return "", errors.New("tx_error") }
    return "0xPENDING_TX", nil
}

// GetComposite returns a stubbed composite (0) for now.
func (c *Client) GetComposite(ctx context.Context) (uint32, error) {
    if !c.cfg.IsEnabled() { return 0, nil }
    // Placeholder: still returns 0 until view call implemented.
    return 0, nil
}

// UpdateWeights validates non-zero sum; returns mock hash.
func (c *Client) UpdateWeights(ctx context.Context, a, g, ml uint32) (string, error) {
    if a+g+ml == 0 { return "", errors.New("zero weights") }
    return "0xMOCK_WTS_TX", nil
}
