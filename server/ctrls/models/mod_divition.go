package models

type Coin struct {
	Owner string
	Value float64
}

type DivitionData struct {
	Coin     string
	NewCoins map[string]Coin
}
