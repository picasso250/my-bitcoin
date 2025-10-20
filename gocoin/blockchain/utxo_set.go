package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"

	"go.etcd.io/bbolt"
)

const utxoBucket = "chainstate"

// UTXOSet represents the UTXO set.
type UTXOSet struct {
	Blockchain *Blockchain
}

// Reindex finds all unspent transaction outputs and rebuilds the UTXO set
func (u UTXOSet) Reindex() {
	db := u.Blockchain.db
	bucketName := []byte(utxoBucket)

	// 1. Delete the old bucket if it exists
	err := db.Update(func(tx *bbolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bbolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// 2. Find all unspent outputs from the entire blockchain
	UTXO := u.Blockchain.FindAllUTXO()

	// 3. Populate the bucket with the found UTXOs
	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key := []byte(txID)
			err := b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// Update updates the UTXO set with the transactions from the new block.
// NOTE: This function should be called AFTER the new block has been added to the DB.
func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.db

	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			// Remove spent outputs (inputs of the new transaction) from the UTXO set
			if !tx.IsCoinbase() {
				for _, vin := range tx.Vin {
					updatedOuts := TxOutputs{}
					outsBytes := b.Get(vin.Txid)
					outs := DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.Txid, updatedOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}

			// Add new unspent outputs from the new transaction
			newOutputs := TxOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := b.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// FindUTXO finds all unspent transaction outputs for a given public key hash.
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	db := u.Blockchain.db

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// TxOutputs serves as a container for serializing an array of TxOutput
type TxOutputs struct {
	Outputs []TxOutput
}

// Serialize serializes TxOutputs
func (outs TxOutputs) Serialize() []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// DeserializeOutputs deserializes TxOutputs
func DeserializeOutputs(data []byte) TxOutputs {
	var outputs TxOutputs
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}
	return outputs
}