package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"

	wallet "blockchain.com/Wallet"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx Transaction) Serialize()[]byte {
	var buff bytes.Buffer
	encoded := gob.NewEncoder(&buff)
	err := encoded.Encode(tx)
	if err!=nil{
		log.Panic(err)
	}
	return buff.Bytes()

	
}
func (tx *Transaction) Hash() []byte{
	var hash [32]byte
	txcopy:=*tx
	txcopy.ID=[]byte{}
	hash=sha256.Sum256(tx.Serialize())
	return hash[:]
}
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	Handle(err)
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}
func CoinBaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coin to %s", to)

	}
	input := TxInput{[]byte{}, -1, nil, []byte(data)}
	output := NewOutput(100, to)
	transaction := &Transaction{nil, []TxInput{input}, []TxOutput{*output}}
	transaction.SetID()
	return transaction

}
func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput
	wallets, err := wallet.CreateWallets()
	Handle(err)
	w := wallets.GetWallet(from)
	pubKeyHash := wallet.PubicKeyHash(w.PublicKey)
	acc, validOutputs := chain.FindSpendableOutputs(pubKeyHash, amount)
	if acc < amount {
		log.Panic("Errors: Not enough funds")
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)
		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
		outputs = append(outputs, *NewOutput(amount, to))
	}
	if acc > amount {
		outputs = append(outputs, *NewOutput(acc-amount, from))
	}
	transaction := Transaction{nil, inputs, outputs}
	transaction.ID = transaction.Hash()
	return &transaction
}
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == -1 && tx.Inputs[0].Out == -1 && len(tx.Inputs[0].ID) == 0
}
func (input *TxInput) CanUnlock(Signature []byte) bool {
	return bytes.Compare(input.Signature,Signature)==0
}
func (output *TxOutput) CanbeUnlocked(pubKeyHash []byte) bool {
	return bytes.Compare(output.PubKeyHash, pubKeyHash) == 0
}
func
func (tx *Transaction)TrimmedCopy() Transaction{
	var inputs []TxInput
	var outputs []TxOutput
	for _,in:=range tx.Inputs{
		inputs = append(inputs, TxInput{in.ID,in.Out,nil,nil})
	}
	for _,out:=range tx.Outputs{
		outputs = append(outputs, TxOutput{out.Value,out.PubKeyHash})
	}
	txcopy:=Transaction{tx.ID,inputs,outputs}
	return txcopy
}