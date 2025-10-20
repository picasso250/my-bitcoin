package gocoin

import (
	"encoding/json"
)

// Message is the top-level envelope for all P2P messages
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Version handshake
type VersionPayload struct {
	Version   int      `json:"version"`   // protocol version
	Agent     string   `json:"agent"`     // user-agent
	Height    int32    `json:"height"`    // current best height
	Services  []string `json:"services"`  // capabilities
	Timestamp int64    `json:"timestamp"` // unix seconds
}

// Verack is empty
type VerackPayload struct{}

// GetBlocks requests block hashes after given locator
type GetBlocksPayload struct {
	Hashes [][]byte `json:"hashes"` // locator hashes, newest first
	Stop   []byte   `json:"stop"`   // optional stop hash (nil=all)
}

// Inv broadcasts inventory
type InvPayload struct {
	Items []InvItem `json:"items"`
}

type InvItem struct {
	Type string `json:"type"` // "block" | "tx"
	Hash []byte `json:"hash"` // 32-byte hash
}

// GetData requests full objects
type GetDataPayload struct {
	Items []InvItem `json:"items"`
}

// Block carries full block
type BlockPayload struct {
	Block JSONBlock `json:"block"`
}

// Tx carries full transaction
type TxPayload struct {
	Tx JSONTx `json:"tx"`
}

// Ping/Pong keepalive
type PingPayload struct {
	Nonce uint64 `json:"nonce"`
}
type PongPayload struct {
	Nonce uint64 `json:"nonce"`
}

type GetMempoolPayload struct{}
type MempoolPayload struct{ Hashes [][]byte `json:"hashes"` }