package main

import (
	"crypto/md5"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"unicode/utf16"
)

// The DNF client validates the login token by applying the RSA public key to
// it. We therefore have to perform the matching private-key operation here.
const privateKeyPEM = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA2Mzieb9pITcA4f8ZlcJ678aFWWwQURecoJCMUdA5Cg5X6w5/
umgvWmOgadWeZKTehHn08KIObx7L0x/4RerkZFYPdigu7fOoJWUle7i3PU2Ussyx
Qa56O5eAQDuQZGSuKKEqUWL7CFpXJUat0kYg8/6o3U9goWcu4jGokjrUTsrj0++9
3mGssPAeB6NY+LOdkVlw4cQKm7EFmxXeMf3gdnD461PKP6XnlVaLliA+mlEeGpqU
HfQ8S4RlnxGXLP47kgCeU2OA0T1MgRGV8PpduEI5G65TVZ7y7jQrNJZRj3ofVV0t
jB7fNqf/Gj/qyOKBYUWVrZGU2btRHejMVkOnIwIDAQABAoIBAQDTZ6aIFbBMJTiF
J54pPMVoPmsV8ZxPlviyUYGi3aphNe9hVHgUqzdRqshnq1iSx3n8MHg6lawBi0Qy
EYClnREtDgZxr2ljuy8BmAnfBRYZfyc62wMWCy9CIM980xhP7SUulUmQpzYmxQEp
ZixlWOEVTAQaGicd/GHpS4cXYKpaUQ4h0RnjOSut72vR80Dpb2mIM9wkWLPoFO5j
dAsZ4t/CAReO8GgnAPWWFJ1tmZ+wNJCbUN9hp38bUYlDErmkihGzAKGtsSlZHPO7
1cJngaVaJW4igFpiAnkevEfb4kwj/OVQZLdYe1P3wLdkCK+QJu/VCfKC3VzTo+ee
clrRDsnxAoGBAPA4jdxJNoxvOK8mabEiRnh79VEfLc3edOZ+rFK6wN7+S5gj37bd
EnEg2pBpum1XohnObvCQ2UOBuDzKDrWg8WEx++HOdmh8gPXeiImPJhqw3zh2yubc
7cfzNmo+jo2OJAwZXT4LNz/NZY1vBlJjbYdVaWnhaEZYEzU4u2ZMpYO3AoGBAOcK
gCipawANkEO7HfCAZxeTx3f1+JCOH08wJL7agRM8ZA5MP6RwwqpqSNYDlOGjINg/
05AOc6rEM+S+4f91O914qU+32Hf1VqAp65N8zfgaiB6yTl6HjvgZ0qoBtfs+xjoF
rgpSd0ONSzJAsTRfwjG4Xsv3K0VSKwIYjpF4wi/1AoGBAKjlOmiNWTr34ATVny4Z
iS6hCOQWZd/+7nY4zfQEtiKS7Jd1cY9ic9ryXHl4vMiv+prmV33webFK/TxQXHM7
SzspJ42l2f6wuwMjiSAec44EmY1biGE1KEGTMMkWXqgMLjALjVFYFWpYgXQlgW5f
1dx3Ivx8rRH4Ttj2oFvWu/CVAoGBALuDdDyMNAGZAItYJJZ7QDCMigm8on6AOr0E
b+5OXRAFpZdNHyIROo0hMfgwc/cldJTOSKDUeeAQ2aU/nyC8P1gihBflOFUz05iu
LTpIwsoojm2LfbuO/eQy4R2FwfMmIbAZyAUspQs4c91XT/sX9P5xis6zznS2mZ1b
3OoIPmK5AoGAAVUNzBGywcFpR86nmOlXbUJu0aTHEg+J5ktrx+JlVyLOKShVvbKm
e53x8Wu2D/ygtmcGJTXZrQV9VqrJrdUqmGuHVVxg7YjaWwmtrF4Lap5ESBv/WtCK
xjx9KoHY0ZrfXwfrMdPsBAsY6wb2QjpbhI7+8lTV6x+yMEVGTEkpgiU=
-----END RSA PRIVATE KEY-----
`

const tokenPrefix = "01FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00"
const tokenSuffix = "010101010101010101010101010101010101010101010101010101010101010155914510010403030101"

var (
	rsaKeyOnce sync.Once
	rsaKeyD    *big.Int
	rsaKeyN    *big.Int
	rsaKeyErr  error
)

func loadPrivateKey() (*big.Int, *big.Int, error) {
	rsaKeyOnce.Do(func() {
		block, _ := pem.Decode([]byte(strings.TrimSpace(privateKeyPEM)))
		if block == nil {
			rsaKeyErr = fmt.Errorf("????????")
			return
		}
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			rsaKeyErr = err
			return
		}
		rsaKeyD, rsaKeyN = key.D, key.N
	})
	return rsaKeyD, rsaKeyN, rsaKeyErr
}

// generateToken mirrors the byte layout used by the original Node client.
func generateToken(uid int) (string, error) {
	d, n, err := loadPrivateKey()
	if err != nil {
		return "", err
	}

	payload, err := hex.DecodeString(fmt.Sprintf("%s%08x%s", tokenPrefix, uint32(uid), tokenSuffix))
	if err != nil {
		return "", err
	}

	m := new(big.Int).SetBytes(payload)
	out := new(big.Int).Exp(m, d, n)
	return base64.StdEncoding.EncodeToString(out.Bytes()), nil
}

// md5Password matches Node's Buffer.from(password, "ascii") behaviour:
// every UTF-16 code unit is truncated to its lowest 8 bits.
func md5Password(password string) string {
	units := utf16.Encode([]rune(password))
	raw := make([]byte, len(units))
	for i, u := range units {
		raw[i] = byte(u)
	}
	sum := md5.Sum(raw)
	return hex.EncodeToString(sum[:])
}
