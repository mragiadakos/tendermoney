package utils

import (
	"encoding/hex"
	"errors"
	"math"

	"github.com/dedis/kyber"

	"github.com/dedis/kyber/group/edwards25519"
	"github.com/dedis/kyber/sign/schnorr"
	"github.com/dedis/kyber/util/key"
)

func CreateKeyPair() (*key.Pair, string) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	kp := key.NewKeyPair(suite)
	b, _ := kp.Public.MarshalBinary()
	out := hex.EncodeToString(b)
	return kp, out
}

func UnmarshalPublicKey(pub string) (kyber.Point, error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	p := suite.Point()
	b, err := hex.DecodeString(pub)
	if err != nil {
		return nil, err
	}
	p.UnmarshalBinary(b)
	return p, nil
}

func Sign(priv kyber.Scalar, msg []byte) (string, error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()

	s, err := schnorr.Sign(suite, priv, msg)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(s), nil
}

func MultiSignature(privs []kyber.Scalar, msg []byte) (string, error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	onePrivate := privs[0]
	for i := 1; i < len(privs); i++ {
		onePrivate = suite.Scalar().Add(onePrivate, privs[i])
	}
	s, err := schnorr.Sign(suite, onePrivate, msg)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(s), nil
}

func MultiVerify(pubHexs []string, sig []byte, msg []byte) (bool, error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	pubs := []kyber.Point{}
	for _, v := range pubHexs {
		kp, err := UnmarshalPublicKey(v)
		if err != nil {
			return false, err
		}
		pubs = append(pubs, kp)
	}
	onePublic := pubs[0]
	for i := 1; i < len(pubs); i++ {
		onePublic = suite.Point().Add(onePublic, pubs[i])
	}
	err := schnorr.Verify(suite, onePublic, msg, sig)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func Verify(pubHex string, sig []byte, msg []byte) (bool, error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	pub := suite.Point()
	pubB, err := hex.DecodeString(pubHex)
	if err != nil {
		return false, errors.New("The public key is not hex: " + err.Error())
	}

	err = pub.UnmarshalBinary(pubB)
	if err != nil {
		return false, errors.New("The public key is not correct.")
	}
	err = schnorr.Verify(suite, pub, msg, sig)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
