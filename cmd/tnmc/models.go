package main

type KeyPairJson struct {
	PublicKey  string
	PrivateKey string
}

type CoinJson struct {
	OwnerPrivateKey string
	OwnerPublicKey  string
	UUID            string
	Value           float64
}
