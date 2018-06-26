package models

import (
	"encoding/hex"
	"errors"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"
)

type ProofVerificationPoints struct {
	G  kyber.Point
	H  kyber.Point
	XG kyber.Point
	XH kyber.Point
}

type ProofVerification struct {
	GHex  string
	HHex  string
	XGHex string
	XHHex string
}

func (pv *ProofVerification) GetProof() (*ProofVerificationPoints, error) {
	if len(pv.GHex) == 0 {
		return nil, errors.New("G is empty")
	}
	if len(pv.HHex) == 0 {
		return nil, errors.New("H is empty")
	}
	if len(pv.XHHex) == 0 {
		return nil, errors.New("XH is empty")
	}
	if len(pv.XGHex) == 0 {
		return nil, errors.New("XG is empty")
	}

	suite := edwards25519.NewBlakeSHA256Ed25519()
	pvp := ProofVerificationPoints{}

	// for G
	gb, err := hex.DecodeString(pv.GHex)
	if err != nil {
		return nil, errors.New("G is not correct hex: " + err.Error())
	}
	pvp.G = suite.Point()
	err = pvp.G.UnmarshalBinary(gb)
	if err != nil {
		return nil, errors.New("G is not correct.")
	}

	// for H
	hb, err := hex.DecodeString(pv.HHex)
	if err != nil {
		return nil, errors.New("H is not correct hex: " + err.Error())
	}
	pvp.H = suite.Point()
	err = pvp.H.UnmarshalBinary(hb)
	if err != nil {
		return nil, errors.New("H is not correct.")
	}

	// for XH
	xhb, err := hex.DecodeString(pv.XHHex)
	if err != nil {
		return nil, errors.New("XH is not correct hex: " + err.Error())
	}
	pvp.XH = suite.Point()
	err = pvp.XH.UnmarshalBinary(xhb)
	if err != nil {
		return nil, errors.New("XH is not correct.")
	}

	// for XG
	xgb, err := hex.DecodeString(pv.XGHex)
	if err != nil {
		return nil, errors.New("XG is not correct hex: " + err.Error())
	}
	pvp.XG = suite.Point()
	err = pvp.XG.UnmarshalBinary(xgb)
	if err != nil {
		return nil, errors.New("XG is not correct.")
	}
	return &pvp, nil
}

func NewProofVerification(g, h, xg, xh kyber.Point) ProofVerification {
	pv := ProofVerification{}

	gb, _ := g.MarshalBinary()
	pv.GHex = hex.EncodeToString(gb)

	hb, _ := h.MarshalBinary()
	pv.HHex = hex.EncodeToString(hb)

	xgb, _ := xg.MarshalBinary()
	pv.XGHex = hex.EncodeToString(xgb)

	xhb, _ := xh.MarshalBinary()
	pv.XHHex = hex.EncodeToString(xhb)

	return pv
}

type ReceiveData struct {
	TransactionHash   string            // sha256 hex
	NewOwners         map[string]string // map[uuid]public_key_hex
	ProofVerification ProofVerification
}
