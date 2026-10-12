package random

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	mRand "math/rand"
)

func GenerateString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func GenerateSecretKey() string {
	return GenerateString(16)
}

func GeneratePluginKey() string {
	return GenerateString(32)
}

func GenerateRandomString(length int) string {

	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[mRand.Intn(len(charset))]
	}
	return string(result)
}

func GeneratePassword(length int) (string, error) {

	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789" +
		"!@#$%^&*()-_=+[]{}<>?/"

	pass := make([]byte, length)

	for i := 0; i < length; i++ {

		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		pass[i] = charset[n.Int64()]
	}

	return string(pass), nil
}
