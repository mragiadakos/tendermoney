package utils

import (
	"encoding/hex"
	"errors"

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

func Verify(pubHex string, sigHex string, msg []byte) (bool, error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	var pub kyber.Point
	pubB, err := hex.DecodeString(pubHex)
	if err != nil {
		return false, errors.New("The public key is not hex: " + err.Error())
	}
	sigB, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, errors.New("The signature is not hex: " + err.Error())
	}
	err = pub.UnmarshalBinary(pubB)
	if err != nil {
		return false, errors.New("The public key is not correct.")
	}
	err = schnorr.Verify(suite, pub, msg, sigB)
	if err != nil {
		return false, nil
	}
	return true, nil
}
