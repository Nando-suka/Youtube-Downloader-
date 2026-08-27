package middleware

import (
	"net/http"
)

// SecurityHeadersMiddleware menambahkan security headers ke semua response
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content Security Policy - ketat untuk mencegah XSS
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data: https:; "+
				"font-src 'self' data:; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none';")

		// X-Frame-Options - mencegah clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options - mencegah MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection - legacy browser protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy - kontrol informasi referrer
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy - kontrol fitur browser
		w.Header().Set("Permissions-Policy",
			"geolocation=(), microphone=(), camera=(), payment=()")

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware mengatur CORS dengan ketat
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Cek apakah origin diizinkan
			allowed := false
			if len(allowedOrigins) == 0 {
				// Jika tidak ada allowed origins, hanya izinkan same-origin
				allowed = origin == ""
			} else {
				for _, allowedOrigin := range allowedOrigins {
					if origin == allowedOrigin {
						allowed = true
						break
					}
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
				w.Header().Set("Access-Control-Max-Age", "3600")
			}

			// Handle preflight request
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
