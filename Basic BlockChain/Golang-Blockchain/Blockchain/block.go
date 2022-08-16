package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash        []byte
	Transaction []*Transaction
	PrevHash    []byte
	Nonce       int
}
func (b *Block) HashTransaction() []byte{
	var txHashes [][]byte
	var txhash [32]byte
	for _,tx:=range b.Transaction{
		txHashes=append(txHashes, tx.ID)
	}
	txhash=sha256.Sum256(bytes.Join(txHashes,[]byte{}))
	return txhash[:]
}

func CreateBlock(txs []*Transaction, prevhash []byte) *Block {
	block := &Block{[]byte{},txs, prevhash, 0}
	pow := NewProof(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}
func Genesis(coinbase *Transaction) *Block {
	genesisBlock := CreateBlock([]*Transaction{coinbase}, []byte{})
	return genesisBlock
}
func (block *Block) Serialize() []byte {
	var res bytes.Buffer
	encode := gob.NewEncoder(&res)
	err := encode.Encode(block)
	Handle(err)
	return res.Bytes()
}
func Deserialize(data []byte) *Block {
	var block Block
	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(&block)
	Handle(err)
	return &block
}
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
