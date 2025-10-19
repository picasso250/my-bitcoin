package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
)

const walletFile = "wallets.dat"

// Wallets stores a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// NewWallets creates Wallets and fills it from a file if it exists
func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile()

	return &wallets, err
}

// CreateWallet adds a Wallet to Wallets
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	addressBytes := wallet.GetAddress()
	address := fmt.Sprintf("%x", addressBytes)

	ws.Wallets[address] = wallet

	return address
}

// GetAddresses returns an array of addresses from the wallet file
func (ws *Wallets) GetAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a Wallet by its address
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}
	
	// Create a temporary, serializable representation of the wallets
	var walletsGob map[string]serializableWallet
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&walletsGob)
	if err != nil {
		log.Panic(err)
	}

	// Reconstruct the full Wallet objects from the serializable data
	ws.Wallets = make(map[string]*Wallet)
	for address, sw := range walletsGob {
		wallet := &Wallet{
			PrivateKey: ecdsa.PrivateKey{
				D: new(big.Int).SetBytes(sw.PrivateKey),
				PublicKey: ecdsa.PublicKey{
					Curve: elliptic.P256(),
					X:     new(big.Int),
					Y:     new(big.Int),
				},
			},
			PublicKey: sw.PublicKey,
		}
		// Reconstruct public key coordinates from the full public key bytes
		halfLen := len(wallet.PublicKey) / 2
		wallet.PrivateKey.PublicKey.X.SetBytes(wallet.PublicKey[:halfLen])
		wallet.PrivateKey.PublicKey.Y.SetBytes(wallet.PublicKey[halfLen:])
		
		ws.Wallets[address] = wallet
	}


	return nil
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile() {
	var content bytes.Buffer

	// Create a temporary, serializable representation of the wallets
	walletsGob := make(map[string]serializableWallet)
	for address, wallet := range ws.Wallets {
		walletsGob[address] = serializableWallet{
			PrivateKey: wallet.PrivateKey.D.Bytes(),
			PublicKey:  wallet.PublicKey,
		}
	}

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(walletsGob)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

// serializableWallet is a simplified struct for gob encoding/decoding
type serializableWallet struct {
	PrivateKey []byte
	PublicKey  []byte
}