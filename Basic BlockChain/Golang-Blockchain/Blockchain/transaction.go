package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	wallet "blockchain.com/Wallet"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx Transaction) Serialize() []byte {
	var buff bytes.Buffer
	encoded := gob.NewEncoder(&buff)
	err := encoded.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()

}
func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txcopy := *tx
	txcopy.ID = []byte{}
	hash = sha256.Sum256(tx.Serialize())
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
func NewTransaction(from, to string, amount int, UTXO *UTXOSet) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput
	wallets, err := wallet.CreateWallets()
	Handle(err)
	w := wallets.GetWallet(from)
	pubKeyHash := wallet.PubicKeyHash(w.PublicKey)
	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)
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
	return bytes.Compare(input.Signature, Signature) == 0
}
func (output *TxOutput) CanbeUnlocked(pubKeyHash []byte) bool {
	return bytes.Compare(output.PubKeyHash, pubKeyHash) == 0
}
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTxs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}
	for _, input := range tx.Inputs {
		if (prevTxs[hex.EncodeToString(input.ID)]).ID == nil {
			log.Panic("ERROR: Previous Transaction is not correct")
		}
	}
	txCopy := tx.TrimmedCopy()
	for inputIndex, input := range txCopy.Inputs {
		preTX := prevTxs[hex.EncodeToString(input.ID)]
		txCopy.Inputs[inputIndex].Signature = nil
		txCopy.Inputs[inputIndex].Pubkey = preTX.Outputs[input.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inputIndex].Pubkey = nil
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		Handle(err)
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Inputs[inputIndex].Signature = signature
	}
}
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput
	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}
	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}
	txcopy := Transaction{tx.ID, inputs, outputs}
	return txcopy
}
func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	for _, in := range tx.Inputs {
		if prevTxs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("ERROR : Previous Transaction is not correct")
		}
	}
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	for inputIndex, input := range txCopy.Inputs {
		prevTX := prevTxs[hex.EncodeToString(input.ID)]
		txCopy.Inputs[inputIndex].Signature = nil
		txCopy.Inputs[inputIndex].Pubkey = prevTX.Outputs[input.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inputIndex].Pubkey = nil
		r := big.Int{}
		s := big.Int{}
		SignLen := len(input.Signature)
		r.SetBytes(input.Signature[:SignLen/2])
		s.SetBytes(input.Signature[SignLen/2:])
		x := big.Int{}
		y := big.Int{}
		PubKeyLen := len(input.Pubkey)
		x.SetBytes(input.Pubkey[:PubKeyLen/2])
		y.SetBytes(input.Pubkey[PubKeyLen/2:])
		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}
	return true
}
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.Pubkey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
