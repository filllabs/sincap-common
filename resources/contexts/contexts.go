// Package contexts provides several utilities and context implementations
package contexts

import "github.com/filllabs/sincap-common/random"

// ContextKey is a special key type for using resource context key
type ContextKey string

// NewContextKey is a special random method for using resource context key 32 byte length
func NewContextKey() ContextKey {
	key := random.GetString()
	return ContextKey(key)
}
