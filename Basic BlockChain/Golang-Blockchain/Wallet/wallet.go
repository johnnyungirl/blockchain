package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func PubicKeyHash(pubkey []byte) []byte {
	publickey := sha256.Sum256(pubkey[:])
	hasher := ripemd160.New()
	_, err := hasher.Write(publickey[:])
	if err != nil {
		log.Panic(err)
	}
	publicRipMD := hasher.Sum(nil)
	return publicRipMD
}
func CheckSum(pubhash []byte) []byte {
	firstSha := sha256.Sum256(pubhash)
	secondSha := sha256.Sum256(firstSha[:])
	return secondSha[:checksumLength]
}

func (w *Wallet) Address() []byte {
	pubHash := PubicKeyHash(w.PublicKey)
	versionHash := append([]byte{version}, pubHash...)
	checksum := CheckSum(pubHash)
	fullHash := append(versionHash, checksum...)
	address := Base58Encode(fullHash)
	return address

}
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pub := append(private.X.Bytes(), private.Y.Bytes()...)
	return *private, pub
}
func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}
