package claims

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"gitlab.com/sincap/sincap-common/crypto"
	"gitlab.com/sincap/sincap-common/logging"
	"go.uber.org/zap"
)

// EncryptedClaims holds the necessary information needed
type EncryptedClaims struct {
	Data      []byte `json:"data,omitempty"`
	Audience  string `json:"aud,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	ID        string `json:"jti,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	NotBefore int64  `json:"nbf,omitempty"`
	Subject   string `json:"sub,omitempty"`
}

// Fill taken a mapclaims and copies values
func (e *EncryptedClaims) Fill(j jwt.MapClaims) {
	if j["data"] != nil {
		data := j["data"].(string)
		dec, _ := base64.StdEncoding.DecodeString(data)
		e.Data = dec
	}
	if j["aud"] != nil {
		e.Audience = j["aud"].(string)
	}
	if j["exp"] != nil {
		e.ExpiresAt = int64(j["exp"].(float64))
	}
	if j["jti"] != nil {
		e.ID = j["jti"].(string)
	}
	if j["iat"] != nil {
		e.IssuedAt = int64(j["iat"].(float64))
	}
	if j["iss"] != nil {
		e.Issuer = j["iss"].(string)
	}
	if j["nbf"] != nil {
		e.NotBefore = int64(j["nbf"].(float64))
	}
	if j["sub"] != nil {
		e.Subject = j["sub"].(string)
	}
}

// Valid validates time based claims "exp, iat, nbf".
// There is no accounting for clock skew.
// As well, if any of the above claims are not in the token, it will still
// be considered a valid claim.
func (e EncryptedClaims) Valid() error {
	vErr := new(jwt.ValidationError)
	now := jwt.TimeFunc().Unix()

	// The claims below are optional, by default, so if they are set to the
	// default value in Go, let's not fail the verification for them.
	if !e.VerifyExpiresAt(now, false) {
		delta := time.Unix(now, 0).Sub(time.Unix(e.ExpiresAt, 0))
		vErr.Inner = fmt.Errorf("token is expired by %v", delta)
		vErr.Errors |= jwt.ValidationErrorExpired
	}

	if !e.VerifyIssuedAt(now, false) {
		vErr.Inner = fmt.Errorf("token used before issued")
		vErr.Errors |= jwt.ValidationErrorIssuedAt
	}

	if !e.VerifyNotBefore(now, false) {
		vErr.Inner = fmt.Errorf("token is not valid yet")
		vErr.Errors |= jwt.ValidationErrorNotValidYet
	}

	// no error
	if vErr.Errors == 0 {
		return nil
	}

	return vErr
}

// VerifyAudience Compares the aud claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (e *EncryptedClaims) VerifyAudience(cmp string, req bool) bool {
	return verifyAud(e.Audience, cmp, req)
}

// VerifyExpiresAt Compares the exp claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (e *EncryptedClaims) VerifyExpiresAt(cmp int64, req bool) bool {
	return verifyExp(e.ExpiresAt, cmp, req)
}

// VerifyIssuedAt Compares the iat claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (e *EncryptedClaims) VerifyIssuedAt(cmp int64, req bool) bool {
	return verifyIat(e.IssuedAt, cmp, req)
}

// VerifyIssuer Compares the iss claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (e *EncryptedClaims) VerifyIssuer(cmp string, req bool) bool {
	return verifyIss(e.Issuer, cmp, req)
}

// VerifyNotBefore Compares the nbf claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (e *EncryptedClaims) VerifyNotBefore(cmp int64, req bool) bool {
	return verifyNbf(e.NotBefore, cmp, req)
}

func verifyAud(aud string, cmp string, required bool) bool {
	if aud == "" {
		return !required
	}
	if subtle.ConstantTimeCompare([]byte(aud), []byte(cmp)) != 0 {
		return true
	}
	return false
}

func verifyExp(exp int64, now int64, required bool) bool {
	return now <= exp
}

func verifyIat(iat int64, now int64, required bool) bool {
	if iat == 0 {
		return !required
	}
	return now >= iat
}

func verifyIss(iss string, cmp string, required bool) bool {
	if iss == "" {
		return !required
	}
	if subtle.ConstantTimeCompare([]byte(iss), []byte(cmp)) != 0 {
		return true
	}
	return false
}

func verifyNbf(nbf int64, now int64, required bool) bool {
	if nbf == 0 {
		return !required
	}
	return now >= nbf
}

// Decrypt reveals all payload and returns as DecryptedClaims
func (e *EncryptedClaims) Decrypt(secret string) (*DecryptedClaims, error) {
	data, err := crypto.Decrypt(e.Data, secret)
	if err != nil {
		logging.Logger.Error("Can not decrypt Claims", zap.Error(err))
		return nil, err
	}
	d := DecryptedClaims{}
	err = json.Unmarshal(data, &d)
	if err != nil {
		logging.Logger.Error("Can not unmarshal Claims", zap.Error(err))
		return nil, err
	}

	return &d, nil
}
