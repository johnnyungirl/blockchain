package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/dgraph-io/badger"
)

var (
	utxoPrefix   = []byte("utxo-")
	prefixlength = len(utxoPrefix)
)

type UTXOSet struct {
	blockchain *BlockChain
}

func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	accumulated := 0
	err := u.blockchain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		item := txn.NewIterator(opts)
		for item.Seek(utxoPrefix); item.ValidForPrefix(utxoPrefix); item.Next() {
			value, err := item.Item().Value()
			Handle(err)
			key := item.Item().Key()
			k := bytes.TrimPrefix(key, utxoPrefix)
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(value)
			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}
		}
		return nil
	})
	Handle(err)
	return accumulated, unspentOuts
}
func (u UTXOSet) FindUTXO(pubkeyHash []byte) []TxOutput {
	var UTXO []TxOutput
	err := u.blockchain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		item := txn.NewIterator(opts)
		for item.Seek(utxoPrefix); item.ValidForPrefix(utxoPrefix); item.Next() {
			itum := item.Item()
			value, err := itum.Value()
			Handle(err)
			outs := DeserializeOutputs(value)
			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) {
					UTXO = append(UTXO, out)
				}
			}

		}
		return nil
	})
	Handle(err)
	return UTXO
}
func (u UTXOSet) ReIndex() {
	u.DeleteByPrefix(utxoPrefix)
	UTXO := u.blockchain.FindUTXO()
	err := u.blockchain.Database.Update(func(txn *badger.Txn) error {
		for txId, outs := range UTXO {
			key, err := hex.DecodeString(txId)
			Handle(err)
			key = append(utxoPrefix, key...)
			err = txn.Set(key, outs.Serialize())
			Handle(err)
		}
		return nil
	})
	Handle(err)
}
func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	deleteKeys := func(KeyForDelete [][]byte) error {
		if err := u.blockchain.Database.Update(func(txn *badger.Txn) error {
			for _, key := range KeyForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}
	collectSize := 100000
	u.blockchain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		item := txn.NewIterator(opts)
		defer item.Close()
		KeyForDelete := make([][]byte, 0, collectSize)
		KeyCollected := 0
		for item.Seek(prefix); item.ValidForPrefix(prefix); item.Next() {
			key := item.Item().KeyCopy(nil)
			KeyForDelete = append(KeyForDelete, key)
			KeyCollected++
			if KeyCollected > collectSize {
				if err := deleteKeys(KeyForDelete); err != nil {
					log.Panic(err)
				}
			}
			KeyForDelete = make([][]byte, 0, collectSize)
			KeyCollected = 0
		}
		if KeyCollected > 0 {
			if err := deleteKeys(KeyForDelete); err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}
func (u UTXOSet) CountTransaction() int {
	counter := 0
	err := u.blockchain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		item := txn.NewIterator(opts)
		for item.Seek(utxoPrefix); item.ValidForPrefix(utxoPrefix); item.Next() {
			counter++
		}
		return nil
	})
	Handle(err)
	return counter
}
func (u *UTXOSet) Update(block *Block) {
	err := u.blockchain.Database.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transaction {
			if tx.IsCoinbase() == false {
				updatedOuts := TxOutputs{}
				for _, in := range tx.Inputs {
					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID)
					Handle(err)
					value, err := item.Value()
					Handle(err)
					outs := DeserializeOutputs(value)

					for outIdx, out := range outs.Outputs {
						if outIdx != in.Out {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}
					if len(updatedOuts.Outputs) == 0 {
						if err := txn.Delete(inID); err != nil {
							log.Panic(err)

						}
					} else {
						if err := txn.Set(inID, updatedOuts.Serialize()); err != nil {
							log.Panic(err)
						}
					}
				}

			}
			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			txID := append(utxoPrefix, tx.ID...)
			if err := txn.Set(txID, newOutputs.Serialize()); err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	Handle(err)

}
