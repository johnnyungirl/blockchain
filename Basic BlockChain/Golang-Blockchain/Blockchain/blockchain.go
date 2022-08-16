package blockchain

import (
	"encoding/hex"
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
func (chain *BlockChain) FindUnspentTransaction(pubKeyHash []byte) []Transaction {
	iter := chain.Interator()
	spentUTXOs := make(map[string][]int)
	var UnspentTxs []Transaction
	for {
		block := iter.Next()
		for _, tx := range block.Transaction {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for OutIdx, Out := range tx.Outputs {
				if spentUTXOs[txID] != nil {
					for _, spentOut := range spentUTXOs[txID] {
						if spentOut == OutIdx {
							continue Outputs
						}
					}
				}
				if Out.CanbeUnlocked(pubKeyHash) {
					UnspentTxs = append(UnspentTxs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, input := range tx.Inputs {
					if input.CanUnlock(pubKeyHash) {
						inTxID := hex.EncodeToString(input.ID)
						spentUTXOs[inTxID] = append(spentUTXOs[inTxID], input.Out)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UnspentTxs
}
func (chain *BlockChain) FindUTXO(pubkeyHash []byte) []TxOutput {
	UnspentTransactions := chain.FindUnspentTransaction(pubkeyHash)
	var UTXOs []TxOutput
	for _, tx := range UnspentTransactions {
		for _, output := range tx.Outputs {
			if output.CanbeUnlocked(pubkeyHash) {
				UTXOs = append(UTXOs, output)
			}
		}
	}
	return UTXOs
}
func (chain *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	accumulated := 0
	UnspentTransactions := chain.FindUnspentTransaction(pubKeyHash)
	SpendableOutPuts := make(map[string][]int)
Work:
	for _, tx := range UnspentTransactions {
		txID := hex.EncodeToString(tx.ID)
		for OutIdx, Output := range tx.Outputs {
			if Output.CanbeUnlocked(pubKeyHash) && accumulated < amount {
				accumulated += amount
				SpendableOutPuts[txID] = append(SpendableOutPuts[txID], OutIdx)
				if accumulated >= amount {
					break Work
				}
			}

		}
	}
	return accumulated, SpendableOutPuts
}
