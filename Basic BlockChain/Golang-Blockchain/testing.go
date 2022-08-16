package main

import (
	"crypto/elliptic"
	"fmt"
)

type Getter interface {
	Get() string
}

type Foo struct {
	Bar string
}

func (f Foo) Get() string {
	return f.Bar
}
func main() {
	curve := elliptic.P256()
	fmt.Println(curve.Params())
}
