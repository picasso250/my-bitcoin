package p2p

// 纯类型，不给方法
type JSONBlock struct {
	Hash          string   `json:"hash"`
	PrevBlockHash string   `json:"prevBlockHash"`
	Timestamp     int64    `json:"timestamp"`
	Nonce         int      `json:"nonce"`
	Transactions  []JSONTx `json:"transactions"`
}

type JSONTx struct {
	ID   string      `json:"id"`
	Vin  []JSONTxIn  `json:"vin"`
	Vout []JSONTxOut `json:"vout"`
}

type JSONTxIn struct {
	Txid      string `json:"txid"`
	Vout      int    `json:"vout"`
	Signature string `json:"signature"` // hex
	PubKey    string `json:"pubKey"`    // hex
}

type JSONTxOut struct {
	Value      int    `json:"value"`
	PubKeyHash string `json:"pubKeyHash"` // hex
}