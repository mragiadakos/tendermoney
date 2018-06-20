package models

var CONSTANT_VALUES = []float64{500, 100, 50, 20, 10, 5, 2, 1, 0.50, 0.20, 0.10, 0.05, 0.02, 0.01}

type InflationData struct {
	Coin     string //uuid
	Value    float64
	Owner    string //public key hex
	Inflator string //public key hex
}
