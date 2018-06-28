package query

import (
	"errors"

	"github.com/mragiadakos/tendermoney/server/ctrls/dbpkg"
)

var (
	ERR_THERE_NO_TAXES = errors.New("There are no taxes to query.")
)

type QueryModelTax struct {
	Percentage int
	Inflator   string
}

func GetLatestTax(s *dbpkg.State) (*QueryModelTax, error) {
	st := s.GetTax()
	if len(st.Inflator) == 0 {
		return nil, ERR_THERE_NO_TAXES
	}
	qmt := QueryModelTax{}
	qmt.Inflator = st.Inflator
	qmt.Percentage = st.Percentage
	return &qmt, nil
}
