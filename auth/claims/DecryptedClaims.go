package claims

import (
	"encoding/json"

	"gitlab.com/sincap/sincap-common/crypto"

	"gitlab.com/sincap/sincap-common/logging"
	"go.uber.org/zap"
)

// DecryptedClaims holds the necessary information needed
type DecryptedClaims struct {
	UserID    uint                   `json:",omitempty"`
	Username  string                 `json:",omitempty"`
	RoleID    uint                   `json:",omitempty"`
	RoleName  string                 `json:",omitempty"`
	UserAgent string                 `json:",omitempty"`
	UserIP    string                 `json:",omitempty"`
	Audience  string                 `json:",omitempty"`
	ExpiresAt int64                  `json:",omitempty"`
	ID        string                 `json:",omitempty"`
	IssuedAt  int64                  `json:",omitempty"`
	Issuer    string                 `json:",omitempty"`
	NotBefore int64                  `json:",omitempty"`
	Subject   string                 `json:",omitempty"`
	Extra     map[string]interface{} `json:",omitempty"`
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

// GetExtra  retuns the extra element with the given key
func (c *DecryptedClaims) GetExtra(key string) (interface{}, bool) {
	val, found := c.Extra[key]
	return val, found
}

// GetExtraNumber  retuns the extra element with the given key as number
func (c *DecryptedClaims) GetExtraNumber(key string) (float64, bool) {
	val, found := c.GetExtra(key)
	if !found {
		return 0, found
	}
	return val.(float64), found
}

// GetExtraString  retuns the extra element with the given key as string
func (c *DecryptedClaims) GetExtraString(key string) (string, bool) {
	val, found := c.GetExtra(key)
	if !found {
		return "", found
	}
	return val.(string), found
}
