package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletfile = "./Database/wallet/wallets.data"

type Wallets struct {
	wallets map[string]*Wallet
}

func (ws *Wallets) Loadfile() error {
	if _, err := os.Stat(walletfile); os.IsNotExist(err) {
		return err
	}
	var purses Wallets
	filecontent, err := ioutil.ReadFile(walletfile)
	if err != nil {
		return err
	}
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(filecontent))
	err = decoder.Decode(&purses)
	if err != nil {
		return err
	}
	ws.wallets = purses.wallets
	return nil
}
func (ws *Wallets) Savefile() {
	var content bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(walletfile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
func CreateWallets() (*Wallets, error) {
	ws := Wallets{}
	ws.wallets = make(map[string]*Wallet)
	err := ws.Loadfile()
	return &ws, err
}
func (ws *Wallets) AddWallet() string {
	wallet := MakeWallet()
	address := fmt.Sprintf("%s", wallet.Address())
	ws.wallets[address] = wallet
	return address
}
func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string
	for address := range ws.wallets {
		addresses = append(addresses, address)

	}
	return addresses

}
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.wallets[address]
}
