package chain

import (
    "os"
    "time"
)

// Config holds Starknet client configuration loaded from environment.
// Real signing / account abstractions intentionally deferred; current implementation is mock/stub.
type Config struct {
    ContractAddress string
    RPCURL          string
    PrivateKey      string
    AccountAddress  string
    Timeout         time.Duration
}

// LoadConfigFromEnv builds Config using environment variables.
// STARKNET_TIMEOUT_MS is optional (default 3000ms).
func LoadConfigFromEnv() Config {
    to := 3 * time.Second
    if v := os.Getenv("STARKNET_TIMEOUT_MS"); v != "" {
        if ms, err := time.ParseDuration(v + "ms"); err == nil {
            to = ms
        }
    }
    return Config{
        ContractAddress: os.Getenv("CONTRACT_ADDRESS"),
        RPCURL:          os.Getenv("STARKNET_RPC"),
        PrivateKey:      os.Getenv("STARKNET_KEY"),
        AccountAddress:  os.Getenv("STARKNET_ACCOUNT"),
        Timeout:         to,
    }
}

// IsEnabled returns true if mandatory fields for real operation are present.
func (c Config) IsEnabled() bool {
    return c.ContractAddress != "" && c.RPCURL != "" && c.PrivateKey != "" && c.AccountAddress != ""
}
