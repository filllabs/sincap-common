package middlewares

import "net/http"

// SecurityHeaders is a middleware for adding security headers to the response
// Cache-Control: no-cache, no-store, max-age=0, must-revalidate
// Pragma: no-cache
// Expires: 0
// X-Content-Type-Options: nosniff
// Strict-Transport-Security: max-age=31536000 ; includeSubDomains
// X-Frame-Options: DENY
// X-XSS-Protection: 1; mode=block
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Add("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
		header.Add("Pragma", "no-cache")
		header.Add("Expires", "0")
		header.Add("X-Content-Type-Options", "nosniff")
		header.Add("Strict-Transport-Security", "max-age=31536000 ; includeSubDomains")
		header.Add("X-Frame-Options", "DENY")
		header.Add("X-XSS-Protection", "1")

		next.ServeHTTP(w, r)
	})
}
