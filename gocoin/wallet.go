package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const addressChecksumLen = 4

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

// GetAddress returns wallet address with version and checksum
func (w Wallet) GetAddress() []byte {
	// 1. Take the public key and hash it twice with SHA-256 and RIPEMD-160
	pubKeyHash := HashPubKey(w.PublicKey)

	// 2. Prepend the version byte to the hash
	versionedPayload := append([]byte{version}, pubKeyHash...)

	// 3. Calculate the checksum by hashing the versioned payload twice with SHA-256
	checksum := checksum(versionedPayload)

	// 4. Append the checksum to the versioned payload
	fullPayload := append(versionedPayload, checksum...)

	return fullPayload
}

// HashPubKey hashes public key with SHA256 and then RIPEMD160
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

// Generates a new ECDSA private-public key pair
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	// 将 X 和 Y 坐标连接起来组成公钥
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}

// DecodeAddress decodes a hex address and returns the public key hash
func DecodeAddress(address string) ([]byte, error) {
	fullPayload, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	if len(fullPayload) < addressChecksumLen+1 {
		return nil, errors.New("invalid address length")
	}

	actualChecksum := fullPayload[len(fullPayload)-addressChecksumLen:]
	versionedPayload := fullPayload[:len(fullPayload)-addressChecksumLen]
	targetChecksum := checksum(versionedPayload)

	if !bytes.Equal(actualChecksum, targetChecksum) {
		return nil, errors.New("invalid address checksum")
	}

	pubKeyHash := versionedPayload[1:] // Remove version byte
	return pubKeyHash, nil
}
