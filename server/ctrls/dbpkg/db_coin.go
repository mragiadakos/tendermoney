package dbpkg

import (
	"encoding/json"
	"errors"
)

var (
	coinKey  = []byte("coin:")
	ownerKey = []byte("owner:")
)

var (
	ERR_COIN_EXISTS_ALREADY = func(uuid string) error {
		return errors.New("The coin " + uuid + " exists already.")
	}
	ERR_COIN_DOES_NOT_EXISTS = func(uuid string) error {
		return errors.New("The coin " + uuid + " does not exists.")
	}

	ERR_OWNER_EXISTS_ALREADY = func(pub string) error {
		return errors.New("The owner " + pub + " exists already.")
	}
	ERR_OWNER_DOES_NOT_EXISTS = func(pub string) error {
		return errors.New("The owner " + pub + " does not exists.")
	}
)

func prefixCoin(uuid string) []byte {
	b := []byte(uuid)
	return append(coinKey, b...)
}

func prefixOwner(pub string) []byte {
	b := []byte(pub)
	return append(ownerKey, b...)
}

type StateCoin struct {
	Coin     string
	Owner    string
	Value    float64
	IsLocked bool
}

func (s *State) AddCoin(sc StateCoin) error {
	has := s.db.Has(prefixCoin(sc.Coin))
	if has {
		return ERR_COIN_EXISTS_ALREADY(sc.Coin)
	}
	has = s.db.Has(prefixOwner(sc.Owner))
	if has {
		return ERR_OWNER_EXISTS_ALREADY(sc.Owner)
	}

	b, _ := json.Marshal(sc)
	s.db.Set(prefixCoin(sc.Coin), b)
	s.db.Set(prefixOwner(sc.Owner), []byte(sc.Coin))
	return nil
}

func (s *State) GetCoin(uuid string) (*StateCoin, error) {
	has := s.db.Has(prefixCoin(uuid))
	if !has {
		return nil, ERR_COIN_DOES_NOT_EXISTS(uuid)
	}
	sc := new(StateCoin)
	b := s.db.Get(prefixCoin(uuid))
	json.Unmarshal(b, &sc)
	return sc, nil
}

func (s *State) DeleteCoinAndOwner(uuid string) {
	sc, err := s.GetCoin(uuid)
	if err != nil {
		return
	}
	s.db.Delete(prefixCoin(uuid))
	s.db.Delete(prefixOwner(sc.Owner))
}

func (s *State) GetOwner(pubHex string) (string, error) {
	has := s.db.Has(prefixOwner(pubHex))
	if !has {
		return "", ERR_OWNER_DOES_NOT_EXISTS(pubHex)
	}
	b := s.db.Get(prefixOwner(pubHex))
	return string(b), nil
}

func (s *State) DeleteOwner(pubHex string) error {
	has := s.db.Has(prefixOwner(pubHex))
	if !has {
		return ERR_OWNER_DOES_NOT_EXISTS(pubHex)
	}
	s.db.Delete(prefixOwner(pubHex))
	return nil
}

func (s *State) SetNewOwner(uuid, owner string) error {
	sc, err := s.GetCoin(uuid)
	if err != nil {
		return err
	}
	sc.Owner = owner
	b, _ := json.Marshal(sc)
	s.db.Set(prefixCoin(sc.Coin), b)
	s.db.Set(prefixOwner(sc.Owner), []byte(sc.Coin))
	return nil
}

func (s *State) IsCoinLocked(uuid string) (bool, error) {
	has := s.db.Has(prefixCoin(uuid))
	if !has {
		return false, ERR_COIN_DOES_NOT_EXISTS(uuid)
	}
	sc := new(StateCoin)
	b := s.db.Get(prefixCoin(uuid))
	json.Unmarshal(b, &sc)
	return sc.IsLocked, nil
}

func (s *State) LockCoin(uuid string) error {
	sc, err := s.GetCoin(uuid)
	if err != nil {
		return err
	}
	sc.IsLocked = true
	b, _ := json.Marshal(sc)
	s.db.Set(prefixCoin(sc.Coin), b)
	return nil
}

func (s *State) UnlockCoin(uuid string) error {
	sc, err := s.GetCoin(uuid)
	if err != nil {
		return err
	}
	sc.IsLocked = false
	b, _ := json.Marshal(sc)
	s.db.Set(prefixCoin(sc.Coin), b)
	return nil
}
