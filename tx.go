package blockchain

import (
	"bytes"

	wallet "blockchain.com/Wallet"
)

type TxInput struct {
	ID        []byte
	Out       int
	Signature []byte
	Pubkey    []byte
}
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

func (input *TxInput) UsesKey(PubKeyHash []byte) bool {
	lockingHash := wallet.PubicKeyHash(input.Pubkey)
	return bytes.Compare(lockingHash, PubKeyHash) == 0

}
func (output *TxOutput) Lock(address []byte) {
	fullpubKeyHash := wallet.Base58Decode(address)
	fullpubKeyHash = fullpubKeyHash[1 : len(fullpubKeyHash)-4]
	output.PubKeyHash = fullpubKeyHash
}
func (output *TxOutput) IsLockedWithKey(pubkeyHash []byte) bool {
	return bytes.Compare(output.PubKeyHash, pubkeyHash) == 0
}
func NewOutput(value int, address string) *TxOutput {
	txo := TxOutput{value, nil}
	txo.Lock([]byte(address))
	return &txo
}
