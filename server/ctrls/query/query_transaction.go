package query

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/mragiadakos/tendermoney/server/ctrls/dbpkg"
)

type QueryModelTransaction struct {
	Hash            string
	Coins           []string
	Fee             []string
	IsFeeReceived   bool
	IsCoinsReceived bool
}

var (
	ERR_TRANSACTION_HAS_NOT_BEEN_SUBMITTED = errors.New("The transaction's hash has not been submitted.")
	ERR_TRANSACTION_HAS_NOT_BEEN_FOUND     = func(hash string) error {
		return errors.New("The transaction's hash has not been found.")
	}
)

func GetTransaction(s *dbpkg.State, u *url.URL) (*QueryModelTransaction, error) {
	values := u.Query()
	hash := values.Get("hash")
	if len(hash) == 0 {
		return nil, ERR_TRANSACTION_HAS_NOT_BEEN_SUBMITTED
	}
	st, err := s.GetTransaction(hash)
	if err != nil {
		return nil, ERR_TRANSACTION_HAS_NOT_BEEN_FOUND(hash)
	}
	qmt := QueryModelTransaction{}
	qmt.Coins = st.Coins
	qmt.Fee = st.Fee
	qmt.Hash = hash
	qmt.IsCoinsReceived = st.IsCoinsReceived
	qmt.IsFeeReceived = st.IsFeeReceived
	return &qmt, nil
}

func GetTransactionsWithUnreceivedFee(s *dbpkg.State) []QueryModelTransaction {
	sts := s.GetTransactions()
	qmts := []QueryModelTransaction{}
	for _, st := range sts {
		if !st.IsFeeReceived {
			qmt := QueryModelTransaction{}
			qmt.Coins = st.Coins
			qmt.Fee = st.Fee

			msg, _ := json.Marshal(st.Coins)
			hash := sha256.Sum256(msg)
			hashHex := hex.EncodeToString(hash[:])

			qmt.Hash = hashHex
			qmt.IsCoinsReceived = st.IsCoinsReceived
			qmt.IsFeeReceived = st.IsFeeReceived
			qmts = append(qmts, qmt)
		}
	}
	return qmts
}
