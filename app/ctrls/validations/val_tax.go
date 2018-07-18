package validations

import (
	"encoding/json"
	"errors"

	"github.com/mragiadakos/tendermoney/app/ctrls/utils"

	"github.com/mragiadakos/tendermoney/app/confs"
	"github.com/mragiadakos/tendermoney/app/ctrls/dbpkg"
	"github.com/mragiadakos/tendermoney/app/ctrls/models"
)

var (
	ERR_TAX_NEGATIVE         = errors.New("The tax can not be negative.")
	ERR_TAX_OVER_ONE_PERCENT = errors.New("The tax can not be over 100.")
)

func ValidateTax(s *dbpkg.State, td models.TaxData, sig []byte) (uint32, error) {
	if td.Percentage < 0 {
		return models.CodeTypeUnauthorized, ERR_TAX_NEGATIVE
	}

	if td.Percentage > 100 {
		return models.CodeTypeUnauthorized, ERR_TAX_OVER_ONE_PERCENT
	}

	inList := false
	for _, v := range confs.Conf.Inflators {
		if v == td.Inflator {
			inList = true
			break
		}
	}
	if !inList {
		return models.CodeTypeUnauthorized, ERR_INFLATOR_NOT_IN_LIST
	}

	msg, _ := json.Marshal(td)
	isVal, err := utils.Verify(td.Inflator, sig, msg)
	if err != nil {
		return models.CodeTypeUnauthorized, err
	}

	if !isVal {
		return models.CodeTypeUnauthorized, ERR_SIGNATURE_NOT_VALID
	}
	return models.CodeTypeOK, nil
}
