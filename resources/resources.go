// Package resources contains the utilities for recources.
package resources

import "net/http"

// Handler is the shorthand type for a fucntion which takes http.Handler and returns http.Handler
type Handler func(next http.Handler) http.Handler
