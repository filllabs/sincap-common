package auth

import (
	"encoding/json"

	"gitlab.com/sincap/sincap-common/crypto"

	"gitlab.com/sincap/sincap-common/logging"
	"go.uber.org/zap"
)

// DecryptedClaims holds the necessary information needed
type DecryptedClaims struct {
	UserID     uint   `json:",omitempty"`
	Username   string `json:",omitempty"`
	RoleID     uint   `json:",omitempty"`
	RoleName   string `json:",omitempty"`
	CurrencyID uint   `json:",omitempty"`
	UserAgent  string `json:",omitempty"`
	UserIP     string `json:",omitempty"`
	Audience   string `json:",omitempty"`
	ExpiresAt  int64  `json:",omitempty"`
	ID         string `json:",omitempty"`
	IssuedAt   int64  `json:",omitempty"`
	Issuer     string `json:",omitempty"`
	NotBefore  int64  `json:",omitempty"`
	Subject    string `json:",omitempty"`
}

// Encrypt hides all payload and returns as EncryptedClaims
func (c *DecryptedClaims) Encrypt(secret string) (*EncryptedClaims, error) {
	data, err := json.Marshal(c)
	if err != nil {
		logging.Logger.Error("Can not marshal DecryptedClaims", zap.Error(err))
		return nil, err
	}
	data, err = crypto.Encrypt(data, secret)
	if err != nil {
		logging.Logger.Error("Can not encrypt DecryptedClaims", zap.Error(err))
		return nil, err
	}
	encoded := EncryptedClaims{
		Audience:  c.Audience,
		ExpiresAt: c.ExpiresAt,
		ID:        c.ID,
		IssuedAt:  c.IssuedAt,
		Issuer:    c.Issuer,
		NotBefore: c.NotBefore,
		Subject:   c.Subject,
		Data:      data,
	}
	return &encoded, nil
}
