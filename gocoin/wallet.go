package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"

	"golang.org/x/crypto/ripemd160"
)

const walletFile = "wallet.json" // 改成 JSON 文件
const version = byte(0x00)
const addressChecksumLen = 4

// Wallet 单钱包结构
type Wallet struct {
	PrivateKey ecdsa.PrivateKey `json:"-"` // 不参与序列化，下面用辅助字段
	PublicKey  []byte           `json:"-"`

	// 明文可读字段
	DHex string `json:"private_key_hex"` // 私钥 D 的 16 进制
	XHex string `json:"public_key_x_hex"`
	YHex string `json:"public_key_y_hex"`
	Pub  string `json:"public_key_concat_hex"` // 拼接后的公钥
}

// LoadOrCreateDefaultWallet 若 wallet.json 不存在则自动创建并返回钱包
func LoadOrCreateDefaultWallet() *Wallet {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		w := NewWallet()
		SaveWallet(w)
		fmt.Printf("New wallet created: %s\n", w.GetAddress())
		return w
	}
	return LoadWallet()
}

// SaveWallet 持久化为明文 JSON
func SaveWallet(w *Wallet) {
	// 把私钥 D 和公钥坐标转成 16 进制
	w.DHex = hex.EncodeToString(w.PrivateKey.D.Bytes())
	w.XHex = hex.EncodeToString(w.PrivateKey.PublicKey.X.Bytes())
	w.YHex = hex.EncodeToString(w.PrivateKey.PublicKey.Y.Bytes())
	w.Pub = hex.EncodeToString(w.PublicKey)

	raw, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		log.Panic(err)
	}
	if err := ioutil.WriteFile(walletFile, raw, 0600); err != nil {
		log.Panic(err)
	}
}

// LoadWallet 从明文 JSON 恢复钱包
func LoadWallet() *Wallet {
	raw, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}
	var aux struct {
		DHex string `json:"private_key_hex"`
		XHex string `json:"public_key_x_hex"`
		YHex string `json:"public_key_y_hex"`
		Pub  string `json:"public_key_concat_hex"`
	}
	if err := json.Unmarshal(raw, &aux); err != nil {
		log.Panic(err)
	}

	// 反解私钥
	curve := elliptic.P256()
	dBytes, _ := hex.DecodeString(aux.DHex)
	xBytes, _ := hex.DecodeString(aux.XHex)
	yBytes, _ := hex.DecodeString(aux.YHex)

	priv := new(ecdsa.PrivateKey)
	priv.Curve = curve
	priv.D = new(big.Int).SetBytes(dBytes)
	priv.PublicKey.X = new(big.Int).SetBytes(xBytes)
	priv.PublicKey.Y = new(big.Int).SetBytes(yBytes)
	priv.PublicKey.Curve = curve

	pub, _ := hex.DecodeString(aux.Pub)

	return &Wallet{
		PrivateKey: *priv,
		PublicKey:  pub,
		DHex:       aux.DHex,
		XHex:       aux.XHex,
		YHex:       aux.YHex,
		Pub:        aux.Pub,
	}
}

// NewWallet 生成新密钥对
func NewWallet() *Wallet {
	private, public := newKeyPair()
	return &Wallet{PrivateKey: private, PublicKey: public}
}

// GetAddress 返回带版本与校验的地址字符串（十六进制）
func (w Wallet) GetAddress() string {
	pubKeyHash := HashPubKey(w.PublicKey)
	versioned := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versioned)
	full := append(versioned, checksum...)
	return hex.EncodeToString(full)
}

/* ---------- 以下逻辑不变 ---------- */

func HashPubKey(pubKey []byte) []byte {
	h1 := sha256.Sum256(pubKey)
	h2 := ripemd160.New()
	h2.Write(h1[:])
	return h2.Sum(nil)
}

func checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:addressChecksumLen]
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(priv.PublicKey.X.Bytes(), priv.PublicKey.Y.Bytes()...)
	return *priv, pubKey
}

// DecodeAddress 解码地址得到公钥哈希
func DecodeAddress(address string) ([]byte, error) {
	full, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	if len(full) < 1+addressChecksumLen {
		return nil, errors.New("invalid address length")
	}
	pubKeyHash := full[1 : len(full)-addressChecksumLen]
	if !bytes.Equal(full[len(full)-addressChecksumLen:], checksum(append([]byte{version}, pubKeyHash...))) {
		return nil, errors.New("invalid checksum")
	}
	return pubKeyHash, nil
}
