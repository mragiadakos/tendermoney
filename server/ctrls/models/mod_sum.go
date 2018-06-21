package models

type SumData struct {
	Coins    []string //uuid
	NewCoin  string   // signature hex
	NewOwner string   //public key hex
}
