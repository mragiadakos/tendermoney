package models

import (
	"encoding/hex"
	"errors"

	"github.com/dedis/kyber/group/edwards25519"
	"github.com/dedis/kyber/proof/dleq"
)

type Proof struct {
	CHex  string
	RHex  string
	VGHex string
	VHHex string
}

func (p *Proof) GetProof() (*dleq.Proof, error) {
	if len(p.CHex) == 0 {
		return nil, errors.New("C is empty")
	}
	if len(p.RHex) == 0 {
		return nil, errors.New("R is empty")
	}
	if len(p.VGHex) == 0 {
		return nil, errors.New("VG is empty")
	}
	if len(p.VHHex) == 0 {
		return nil, errors.New("VH is empty")
	}
	proof := new(dleq.Proof)
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// for C
	cb, err := hex.DecodeString(p.CHex)
	if err != nil {
		return nil, errors.New("C is not correct hex: " + err.Error())
	}
	proof.C = suite.Scalar()
	err = proof.C.UnmarshalBinary(cb)
	if err != nil {
		return nil, errors.New("C is not correct.")
	}

	// for R
	rb, err := hex.DecodeString(p.RHex)
	if err != nil {
		return nil, errors.New("R is not correct hex: " + err.Error())
	}
	proof.R = suite.Scalar()
	err = proof.R.UnmarshalBinary(rb)
	if err != nil {
		return nil, errors.New("R is not correct.")
	}

	// for VG
	vgb, err := hex.DecodeString(p.VGHex)
	if err != nil {
		return nil, errors.New("VG is not correct hex: " + err.Error())
	}
	proof.VG = suite.Point()
	err = proof.VG.UnmarshalBinary(vgb)
	if err != nil {
		return nil, errors.New("VG is not correct.")
	}

	// for VH
	vhb, err := hex.DecodeString(p.VHHex)
	if err != nil {
		return nil, errors.New("VH is not correct hex: " + err.Error())
	}
	proof.VH = suite.Point()
	err = proof.VH.UnmarshalBinary(vhb)
	if err != nil {
		return nil, errors.New("VH is not correct.")
	}
	return proof, nil
}

func NewProof(proof *dleq.Proof) Proof {
	p := Proof{}

	cb, _ := proof.C.MarshalBinary()
	p.CHex = hex.EncodeToString(cb)

	rb, _ := proof.R.MarshalBinary()
	p.RHex = hex.EncodeToString(rb)

	vgb, _ := proof.VG.MarshalBinary()
	p.VGHex = hex.EncodeToString(vgb)

	vhb, _ := proof.VH.MarshalBinary()
	p.VHHex = hex.EncodeToString(vhb)

	return p
}

type SendData struct {
	Coins []string
	Fee   []string
	Proof Proof
}
