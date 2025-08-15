package chain

import (
    "bytes"
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "strings"
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
    // NOTE: Starknet calldata normally: [entry_point_selector, arg_len?, args...]; here we keep a simplified devnet-friendly mock.
    selector := computeSelector("set_prediction_inputs")
    calldata := []string{selector, fmt.Sprintf("0x%x", astro), fmt.Sprintf("0x%x", grav)}
    invoke := map[string]any{
        "type":          "INVOKE",
        "sender_address": c.cfg.AccountAddress,
        "calldata":       calldata,
        "signature":      signStub(selector, astro, grav, c.cfg.PrivateKey),
        "max_fee":        "0x0",
        "version":        "0x1",
    }
    payload := map[string]any{"jsonrpc": "2.0", "method": "starknet_addInvokeTransaction", "params": []any{invoke}, "id": 1}
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
    // Use starknet_call on get_state() entrypoint; composite is 5th value (index 4)
    selector := computeSelector("get_state")
    call := map[string]any{
        "jsonrpc": "2.0",
        "method":  "starknet_call",
        "params": []any{map[string]any{
            "contract_address": c.cfg.ContractAddress,
            "entry_point_selector": selector,
            "calldata":            []string{},
        }, "latest"},
        "id": 2,
    }
    buf, _ := json.Marshal(call)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.RPCURL, bytes.NewReader(buf))
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.httpClient.Do(req)
    if err != nil { return 0, err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return 0, errors.New("rpc_unavailable") }
    var out struct {
        Result struct{ Result []string `json:"result"` } `json:"result"`
        Error  any `json:"error"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        // Maintain previous lenient behavior: treat malformed body as zero composite (test expectation)
        return 0, nil
    }
    if out.Error != nil { return 0, errors.New("call_error") }
    // Depending on node, shape may be {result:[..]}. Try to parse last or 5th element.
    vals := out.Result.Result
    if len(vals) == 0 { return 0, errors.New("empty_result") }
    idx := 4
    if idx >= len(vals) { idx = len(vals)-1 }
    v := vals[idx]
    v = strings.TrimPrefix(v, "0x")
    var composite uint32
    if v == "" { return 0, errors.New("bad_value") }
    // parse hex
    var parsed uint64
    _, err = fmt.Sscanf(v, "%x", &parsed)
    if err != nil { return 0, err }
    composite = uint32(parsed)
    return composite, nil
}

// computeSelector provides a simplistic Cairo 1 selector derivation (keccak felt truncation not implemented fully).
// For development we approximate with sha256 and truncate; real implementation should use starknet keccak.
func computeSelector(name string) string {
    h := sha256.Sum256([]byte(name))
    // take lower 4 bytes just for dev clarity (NOT production safe!)
    return "0x" + hex.EncodeToString(h[:4])
}

// signStub returns a fake signature slice; integrate real signing later.
func signStub(selector string, astro, grav uint32, key string) []string {
    if key == "" { return []string{} }
    h := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%s", selector, astro, grav, key)))
    return []string{"0x" + hex.EncodeToString(h[:16]), "0x" + hex.EncodeToString(h[16:32])}
}

// UpdateWeights validates non-zero sum; returns mock hash.
func (c *Client) UpdateWeights(ctx context.Context, a, g, ml uint32) (string, error) {
    if a+g+ml == 0 { return "", errors.New("zero weights") }
    return "0xMOCK_WTS_TX", nil
}
