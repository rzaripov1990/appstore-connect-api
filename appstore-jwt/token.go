package appstore_jwt

import (
	"appstore-connect-api/config"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func CreateToken(cfg *config.Config) (string, error) {
	headers := map[string]interface{}{
		"alg": "ES256",
		"kid": cfg.Kid,
		"typ": "JWT",
	}

	now := time.Now()
	expirationTime := now.Add(19 * time.Minute)
	block, _ := pem.Decode([]byte(cfg.P8Key))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	signingKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": cfg.Iis,
		"iat": now.Unix(),
		"exp": expirationTime.Unix(),
		"aud": "appstoreconnect-v1",
	})
	token.Header = headers

	signedToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
