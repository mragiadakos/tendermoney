package query

import (
	"errors"
	"net/url"

	"github.com/mragiadakos/tendermoney/app/ctrls/dbpkg"
)

type QueryModelCoin struct {
	Coin     string
	Owner    string
	IsLocked bool
	Value    float64
}

var (
	ERR_COIN_NOT_FOUND = func(uuid string) error {
		return errors.New("The coin " + uuid + " has not been found.")
	}
	ERR_COIN_HAS_NOT_BEEN_SUBMITTED  = errors.New("The coin's UUID has not been submitted.")
	ERR_OWNER_HAS_NOT_BEEN_SUBMITTED = errors.New("The owner's public key has not been submitted.")
	ERR_OWNER_HAS_NOT_BEEN_FOUND     = func(owner string) error {
		return errors.New("The owner " + owner + " has not been found")
	}
)

func GetCoin(s *dbpkg.State, u *url.URL) (*QueryModelCoin, error) {
	values := u.Query()
	uuid := values.Get("coin")
	if len(uuid) == 0 {
		return nil, ERR_COIN_HAS_NOT_BEEN_SUBMITTED
	}
	sc, err := s.GetCoin(uuid)
	if err != nil {
		return nil, ERR_COIN_NOT_FOUND(uuid)
	}

	qm := QueryModelCoin{}
	qm.Coin = uuid
	qm.Owner = sc.Owner
	qm.Value = sc.Value
	qm.IsLocked = sc.IsLocked
	return &qm, nil
}

func GetCoinByOwner(s *dbpkg.State, u *url.URL) (*QueryModelCoin, error) {
	values := u.Query()
	owner := values.Get("owner")
	if len(owner) == 0 {
		return nil, ERR_OWNER_HAS_NOT_BEEN_SUBMITTED
	}

	coin, err := s.GetOwner(owner)
	if err != nil {
		return nil, ERR_OWNER_HAS_NOT_BEEN_FOUND(owner)
	}

	sc, err := s.GetCoin(coin)
	if err != nil {
		return nil, ERR_COIN_NOT_FOUND(coin)
	}

	qm := QueryModelCoin{}
	qm.Coin = coin
	qm.Owner = sc.Owner
	qm.Value = sc.Value
	qm.IsLocked = sc.IsLocked

	return &qm, nil
}
