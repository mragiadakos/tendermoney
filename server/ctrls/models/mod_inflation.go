package models

type InflationData struct {
	Coin     string //uuid
	Value    float64
	Owner    string //public key hex
	Inflator string //public key hex
}
