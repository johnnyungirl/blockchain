package blockchain

import (
	"bytes"
	"encoding/gob"

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
type TxOutputs struct{
	Outputs []TxOutput
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
func (outputs TxOutputs) Serialize() []byte{
	var buff bytes.Buffer
	encoder:=gob.NewEncoder(&buff)
	err:=encoder.Encode(outputs)
	Handle(err)
	return buff.Bytes()
}
func DeserializeOutputs(data []byte) TxOutputs{
	var Outs TxOutputs
	decode:=gob.NewDecoder(bytes.NewReader(data))
	err:=decode.Decode(&Outs)
	Handle(err)
	return Outs
}
