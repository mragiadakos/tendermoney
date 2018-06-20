package dbpkg

import (
	"encoding/json"
	"errors"

	"github.com/mragiadakos/tendermoney/server/ctrls/models"
)

var (
	coinKey  = []byte("coin:")
	ownerKey = []byte("owner:")
)

var (
	ERR_COIN_EXISTS_ALREADY  = errors.New("The coin exists already.")
	ERR_OWNER_EXISTS_ALREADY = errors.New("The owner exists already.")
)

func prefixCoin(uuid string) []byte {
	b := []byte(uuid)
	return append(coinKey, b...)
}

func prefixOwner(pub string) []byte {
	b := []byte(pub)
	return append(ownerKey, b...)
}

func (s *State) AddCoin(id *models.InflationData) error {
	has := s.db.Has(prefixCoin(id.Coin))
	if has {
		return ERR_COIN_EXISTS_ALREADY
	}
	has = s.db.Has(prefixOwner(id.Owner))
	if has {
		return ERR_OWNER_EXISTS_ALREADY
	}
	b, _ := json.Marshal(id)
	s.db.Set(prefixCoin(id.Coin), b)
	s.db.Set(prefixOwner(id.Owner), []byte(id.Coin))
	return nil
}
