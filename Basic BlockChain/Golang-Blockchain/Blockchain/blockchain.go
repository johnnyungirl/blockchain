package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbpath      = "./Database/blocks"
	dbfile      = "./Database/blocks/MAINFEST"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func DBexists() bool {
	if _, err := os.Stat(dbfile); os.IsNotExist(err) {
		return false
	}
	return true
}
func InitBlockchain(address string) *BlockChain {
	if DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}
	var lastHash []byte
	opts := badger.DefaultOptions(dbpath)
	opts.Dir = dbpath
	opts.ValueDir = dbpath
	db, err := badger.Open(opts)
	Handle(err)
	err = db.Update(func(txn *badger.Txn) error {
		coinbase := CoinBaseTx(address, genesisData)
		genesis := Genesis(coinbase)
		fmt.Println("Genesis Created")

		err = txn.Set(lastHash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), lastHash)
		lastHash = genesis.Hash
		return err

	})
	Handle(err)
	blockchain := &BlockChain{lastHash, db}
	return blockchain
}
func ContinueBlockchain(address string) *BlockChain {
	var lastHash []byte
	if DBexists() == false {
		fmt.Println("No existing Blockchain found,create one!")
		runtime.Goexit()
	}
	opts := badger.DefaultOptions(dbpath)
	opts.Dir = dbpath
	opts.ValueDir = dbpath
	db, err := badger.Open(opts)
	Handle(err)
	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()
		return err

	})
	Handle(err)
	blockchain := &BlockChain{lastHash, db}
	return blockchain

}
func (chain *BlockChain) AddBlock(transaction []*Transaction) {
	var lastHash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()
		return err
	})
	Handle(err)

	newblock := CreateBlock(transaction, lastHash)
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newblock.Hash, newblock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newblock.Hash)
		chain.LastHash = newblock.Hash
		return err
	})
	Handle(err)
}
func (chain *BlockChain) Interator() *BlockChainIterator {
	iter := BlockChainIterator{chain.LastHash, chain.Database}
	return &iter
}
func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.Database.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodeBlock, err := item.Value()
		block = Deserialize(encodeBlock)
		return err
	})
	Handle(err)
	iter.CurrentHash = block.PrevHash
	return block
}
func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)
	iter := chain.Interator()
	for {
		block := iter.Next()
		for _, tx := range block.Transaction {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs

				if tx.IsCoinbase() == false {
					for _, input := range tx.Inputs {
						inID := hex.EncodeToString(input.ID)
						spentTXOs[inID] = append(spentTXOs[inID], input.Out)
					}
				}
				if len(block.PrevHash) == 0 {
					break
				}
			}
		}
		return UTXO
	}
}
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Interator()
	for {
		block := iter.Next()
		for _, tx := range block.Transaction {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction dose not exist")
}
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	var prevTxs map[string]Transaction
	for _, input := range tx.Inputs {
		prevTX, err := bc.FindTransaction(input.ID)
		Handle(err)
		prevTxIndex := hex.EncodeToString(prevTX.ID)
		prevTxs[prevTxIndex] = prevTX
	}
	tx.Sign(privKey, prevTxs)
}
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTxs := make(map[string]Transaction)
	for _, input := range tx.Inputs {
		prevTx, err := bc.FindTransaction(input.ID)
		Handle(err)
		prevTxIndex := hex.EncodeToString(prevTx.ID)
		prevTxs[prevTxIndex] = prevTx
	}
	return tx.Verify(prevTxs)
}
