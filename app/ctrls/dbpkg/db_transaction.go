package dbpkg

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/mragiadakos/tendermoney/server/ctrls/models"
)

var (
	ERR_TRANSACTION_NOT_EXIST = func(hash string) error {
		return errors.New("The transaction " + hash + " does not exists.")
	}
)
var (
	transactionKey = []byte("transaction:")
)

func prefixTransaction(hashHex string) []byte {

	return append(transactionKey, []byte(hashHex)...)
}

type StateTransaction struct {
	models.SendData
	IsCoinsReceived bool // the coins retrieved by the receiver
	IsFeeReceived   bool // the fee retrieved by the inflator
}

func (s *State) AddTransaction(sd models.SendData) error {
	st := StateTransaction{}
	st.SendData = sd
	sdb, _ := json.Marshal(st)
	coinb, _ := json.Marshal(sd.Coins)
	hash := sha256.Sum256(coinb)
	hashHex := hex.EncodeToString(hash[:])
	s.db.Set(prefixTransaction(hashHex), sdb)
	return nil
}

func (s *State) GetTransaction(hash string) (*StateTransaction, error) {
	has := s.db.Has(prefixTransaction(hash))
	if !has {
		return nil, ERR_TRANSACTION_NOT_EXIST(hash)
	}
	b := s.db.Get(prefixTransaction(hash))
	st := new(StateTransaction)
	json.Unmarshal(b, &st)
	return st, nil
}

func (s *State) CoinsReceivedFromTransaction(hash string) error {
	st, err := s.GetTransaction(hash)
	if err != nil {
		return err
	}
	st.IsCoinsReceived = true
	stb, _ := json.Marshal(st)
	s.db.Set(prefixTransaction(hash), stb)
	return nil
}

func (s *State) FeeRetrievedFromTransaction(hash string) error {
	st, err := s.GetTransaction(hash)
	if err != nil {
		return err
	}
	st.IsFeeReceived = true
	stb, _ := json.Marshal(st)
	s.db.Set(prefixTransaction(hash), stb)
	return nil
}

func (s *State) GetTransactions() []StateTransaction {
	iter := s.db.Iterator(nil, nil)
	sts := []StateTransaction{}
	for {
		if !iter.Valid() {
			break
		}
		if strings.HasPrefix(string(iter.Key()), string(transactionKey)) {
			st := StateTransaction{}
			json.Unmarshal(iter.Value(), &st)
			sts = append(sts, st)
		}
		iter.Next()
	}
	return sts
}
