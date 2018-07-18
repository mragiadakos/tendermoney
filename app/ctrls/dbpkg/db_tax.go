package dbpkg

import (
	"encoding/json"

	"github.com/mragiadakos/tendermoney/server/ctrls/models"
)

var (
	latestTax = []byte("latestTax")
)

func (s *State) AddTax(tax models.TaxData) {
	b, _ := json.Marshal(tax)
	s.db.Set(latestTax, b)
}

func (s *State) GetTax() models.TaxData {
	has := s.db.Has(latestTax)
	if !has {
		return models.TaxData{}
	}
	b := s.db.Get(latestTax)
	td := models.TaxData{}
	json.Unmarshal(b, &td)
	return td
}
