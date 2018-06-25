package models

import "github.com/mragiadakos/tendermoney/server/ctrls/utils"

type TaxData struct {
	Percentage int
	Inflator   string
}

func (td *TaxData) GetFeeFromTransaction(tr float64) float64 {
	if td.Percentage == 0 {
		return 0
	}
	tax := float64(td.Percentage) / 100.0
	fee := tr * tax
	fixed := utils.ToFixed(fee, 2)
	if fixed == 0 {
		return 0.01
	}

	return fixed
}
