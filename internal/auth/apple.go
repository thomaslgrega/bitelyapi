package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type AppleClaims struct {
	Email     *string `json:"email,omitempty"`
	IsPrivate *bool   `json:"is_private_email,omitempty"`

	jwt.RegisteredClaims
}

type ApplePublicKeys struct {
	Keys []ApplePublicKey `json:"keys"`
}

type ApplePublicKey struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func appleKeyToPublicKey(key ApplePublicKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, err
	}

	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}

func VerifyAppleIdentityToken(tokenString string, bundleId string) (*AppleClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&AppleClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("unexpected signing method")
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, errors.New("missing kid")
			}

			resp, err := http.Get("https://appleid.apple.com/auth/keys")
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			var keys ApplePublicKeys
			if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
				return nil, err
			}

			for _, key := range keys.Keys {
				if key.Kid == kid {
					return appleKeyToPublicKey(key)
				}
			}

			return nil, errors.New("matching Apple public key not found")
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AppleClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Issuer != "https://appleid.apple.com" {
		return nil, errors.New("invalid issuer")
	}

	aud, err := claims.GetAudience()
	if err != nil {
		return nil, errors.New("invalid audience claim")
	}

	validAud := false
	for _, a := range aud {
		if a == bundleId {
			validAud = true
			break
		}
	}

	if !validAud {
		return nil, errors.New("invalid audience")
	}
	
	return claims, nil
}