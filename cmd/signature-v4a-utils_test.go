package cmd

import (
	"encoding/hex"
	"testing"
)

func BenchmarkVerifyV4ASignature(b *testing.B) {
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	hash, _ := hex.DecodeString("936bed871590d4ddd183ed1e8f55bd9fd04c59ac83a1b29953d260f16de30f73")
	sign, _ := hex.DecodeString("3045022056fb6825eea50293f60e6ab56bd9f35f35b6d7570b6493860c625a1794ff9437022100e03491e568b38d95cf7106bbf68cc13f9dfa5c05b7e74747bda0c9047159781c")

	for i := 0; i < b.N; i++ {
		verifyV4ASignature(accessKey, secretKey, hash, sign)
	}
}

func BenchmarkVerifyECDSA(b *testing.B) {
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	hash, _ := hex.DecodeString("936bed871590d4ddd183ed1e8f55bd9fd04c59ac83a1b29953d260f16de30f73")
	sign, _ := hex.DecodeString("3045022056fb6825eea50293f60e6ab56bd9f35f35b6d7570b6493860c625a1794ff9437022100e03491e568b38d95cf7106bbf68cc13f9dfa5c05b7e74747bda0c9047159781c")

	privKey, _ := deriveKeyFromAccessKeyPair(accessKey, secretKey)

	for i := 0; i < b.N; i++ {
		verifyECDSA(&privKey.PublicKey, hash, sign)
	}
}
