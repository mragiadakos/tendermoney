package models

import (
	"math"
)

type TaxData struct {
	Percentage int
	Inflator   string
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func (td *TaxData) GetFeeFromTransaction(tr float64) float64 {
	if td.Percentage == 0 {
		return 0
	}
	tax := float64(td.Percentage) / 100.0
	fee := tr * tax
	fixed := toFixed(fee, 2)
	if fixed == 0 {
		return 0.01
	}

	return fixed
}
