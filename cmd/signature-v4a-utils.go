package cmd

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"math/big"

	"github.com/minio/minio/internal/hash/sha256"
)

var (
	p256               elliptic.Curve
	nMinusTwoP256, one *big.Int
)

func init() {
	// Ensure the elliptic curve parameters are initialized on package import rather then on first usage
	p256 = elliptic.P256()

	one = new(big.Int).SetInt64(1)
	nMinusTwoP256 = new(big.Int).SetBytes(p256.Params().N.Bytes())
	nMinusTwoP256 = nMinusTwoP256.Sub(nMinusTwoP256, new(big.Int).SetInt64(2))
}

func makeHash(hash hash.Hash, b []byte) []byte {
	hash.Reset()
	hash.Write(b)
	return hash.Sum(nil)
}

// constantTimeByteCompare is a constant-time byte comparison of x and y. This function performs an absolute comparison
// if the two byte slices assuming they represent a big-endian number.
//
// 	 error if len(x) != len(y)
//   -1 if x <  y
//    0 if x == y
//   +1 if x >  y
func constantTimeByteCompare(x, y []byte) (int, error) {
	if len(x) != len(y) {
		return 0, fmt.Errorf("slice lengths do not match")
	}

	xLarger, yLarger := 0, 0

	for i := 0; i < len(x); i++ {
		xByte, yByte := int(x[i]), int(y[i])

		x := ((yByte - xByte) >> 8) & 1
		y := ((xByte - yByte) >> 8) & 1

		xLarger |= x &^ yLarger
		yLarger |= y &^ xLarger
	}

	return xLarger - yLarger, nil
}

// verifySignature takes the provided public key, hash, and asn1 encoded signature and returns
// whether the given signature is valid.
func verifyV4ASignature(accessKey, secretKey string, hash []byte, signature []byte) (bool, error) {
	privKey, err := deriveKeyFromAccessKeyPair(accessKey, secretKey)
	if err != nil {
		return false, err
	}
	return verifyECDSA(&privKey.PublicKey, hash, signature)
}

// hmackeyDerivation provides an implementation of a NIST-800-108 of a KDF (Key Derivation Function) in Counter Mode.
// For the purposes of this implantation HMAC is used as the PRF (Pseudorandom function), where the value of
// `r` is defined as a 4 byte counter.
func hmacKeyDerivation(hash func() hash.Hash, bitLen int, key []byte, label, context []byte) ([]byte, error) {
	// verify that we won't overflow the counter
	n := int64(math.Ceil((float64(bitLen) / 8) / float64(hash().Size())))
	if n > 0x7FFFFFFF {
		return nil, fmt.Errorf("unable to derive key of size %d using 32-bit counter", bitLen)
	}

	// verify the requested bit length is not larger then the length encoding size
	if int64(bitLen) > 0x7FFFFFFF {
		return nil, fmt.Errorf("bitLen is greater than 32-bits")
	}

	fixedInput := bytes.NewBuffer(nil)
	fixedInput.Write(label)
	fixedInput.WriteByte(0x00)
	fixedInput.Write(context)
	if err := binary.Write(fixedInput, binary.BigEndian, int32(bitLen)); err != nil {
		return nil, fmt.Errorf("failed to write bit length to fixed input string: %v", err)
	}

	var output []byte

	h := hmac.New(hash, key)

	for i := int64(1); i <= n; i++ {
		h.Reset()
		if err := binary.Write(h, binary.BigEndian, int32(i)); err != nil {
			return nil, err
		}
		_, err := h.Write(fixedInput.Bytes())
		if err != nil {
			return nil, err
		}
		output = append(output, h.Sum(nil)...)
	}

	return output[:bitLen/8], nil
}

// deriveKeyFromAccessKeyPair derives a NIST P-256 PrivateKey from the given
// IAM AccessKey and SecretKey pair.
//
// Based on FIPS.186-4 Appendix B.4.2
func deriveKeyFromAccessKeyPair(accessKey, secretKey string) (*ecdsa.PrivateKey, error) {
	params := p256.Params()
	bitLen := params.BitSize // Testing random candidates does not require an additional 64 bits
	counter := 0x01

	buffer := make([]byte, 1+len(accessKey)) // 1 byte counter + len(accessKey)
	kdfContext := bytes.NewBuffer(buffer)

	inputKey := append([]byte("AWS4A"), []byte(secretKey)...)

	d := new(big.Int)
	for {
		kdfContext.Reset()
		kdfContext.WriteString(accessKey)
		kdfContext.WriteByte(byte(counter))

		key, err := hmacKeyDerivation(sha256.New, bitLen, inputKey, []byte("AWS4-ECDSA-P256-SHA256"), kdfContext.Bytes())
		if err != nil {
			return nil, err
		}

		// Check key first before calling SetBytes if key key is in fact a valid candidate.
		// This ensures the byte slice is the correct length (32-bytes) to compare in constant-time
		cmp, err := constantTimeByteCompare(key, nMinusTwoP256.Bytes())
		if err != nil {
			return nil, err
		}
		if cmp == -1 {
			d.SetBytes(key)
			break
		}

		counter++
		if counter > 0xFF {
			return nil, fmt.Errorf("exhausted single byte external counter")
		}
	}
	d = d.Add(d, one)

	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = p256
	priv.D = d
	priv.PublicKey.X, priv.PublicKey.Y = p256.ScalarBaseMult(d.Bytes())

	return priv, nil
}

type ecdsaSignature struct {
	R, S *big.Int
}

func verifyECDSA(pubKey *ecdsa.PublicKey, hash []byte, signature []byte) (bool, error) {
	var ecdsaSignature ecdsaSignature
	_, err := asn1.Unmarshal(signature, &ecdsaSignature)
	if err != nil {
		return false, err
	}

	return ecdsa.Verify(pubKey, hash, ecdsaSignature.R, ecdsaSignature.S), nil
}
