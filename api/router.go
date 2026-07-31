package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/JohnnyKahiu/speedsales_login/internal/authentication"
	limiter "github.com/JohnnyKahiu/speedsales_login/pkg/Limiter"
	"github.com/gorilla/mux"
)

var mySigningKey = []byte(os.Getenv("SPEEDSALESJWTKEY"))

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
}

func NewRouter() *mux.Router {
	// rentals.CreateTables()

	r := mux.NewRouter()

	// r.HandleFunc("/ws", socketHandler)
	r.Use(RateLimiterMiddleware)

	r.HandleFunc("/login", loginRoutes).Methods("POST", "OPTIONS")
	r.HandleFunc("/login/reset", ResetUserRoutes).Methods("POST", "OPTIONS")

	r.HandleFunc("/users/{module}", UserPOSTRoutes).Methods("POST", "OPTIONS")
	r.HandleFunc("/users/{module}", UserGETRoutes).Methods("GET", "OPTIONS")

	r.HandleFunc("/license", LicensePOST).Methods("POST", "OPTIONS")
	r.HandleFunc("/license", LicenseGET).Methods("GET", "OPTIONS")

	// r.HandleFunc("/sms", sms.Post).Methods("POST", "OPTIONS")

	return r
}

// clientIP extracts the caller's address, preferring X-Forwarded-For (set by
// a reverse proxy) and falling back to the raw connection address. Used
// whenever a request carries no better identifier (no auth token, no
// username in the body), so unrelated anonymous callers don't end up
// sharing one rate-limit bucket keyed on an empty string.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if i := strings.Index(fwd, ","); i != -1 {
			return strings.TrimSpace(fwd[:i])
		}
		return strings.TrimSpace(fwd)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func RateLimiterMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		EnableCors(&w)

		lim := limiter.Limiter{Path: r.URL.Path}
		ctx := r.Context()

		switch lim.Path {
		case "/login":
			b, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body.Close()
			// Restore the body so the handler downstream can still read it
			// directly if it ever needs to — POST currently reads the
			// parsed fields back out of the context instead (see
			// authentication.ParamsContextKey), but nothing should require
			// draining the body here to make that the only option.
			r.Body = io.NopCloser(bytes.NewReader(b))

			var reqArg map[string]string
			if err = json.Unmarshal(b, &reqArg); err != nil {
				log.Printf("error. failed to unmarshal json    err = %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			lim.TokenID = reqArg["username"]
			if lim.TokenID == "" {
				lim.TokenID = clientIP(r)
			}
			ctx = context.WithValue(ctx, authentication.ParamsContextKey, reqArg)

		case "/login/reset":
			// Pre-auth, brute-force-sensitive like /login, but the request
			// shape here isn't owned by this middleware — key on the caller
			// instead of risking a fragile re-parse of a body it doesn't
			// otherwise need to touch.
			lim.TokenID = clientIP(r)

		default:
			// Every other route shares one generic bucket per caller,
			// same as before — just no longer collapsing to a single
			// shared bucket when the token header is missing.
			lim.Path = "/"
			lim.TokenID = r.Header.Get("token")
			if lim.TokenID == "" {
				lim.TokenID = clientIP(r)
			}
		}

		allowed, err := lim.Check(r.Context())
		if err != nil && !errors.Is(err, limiter.ErrLimitReached) {
			log.Printf("rate limiter: allowing request through despite error   path=%s err=%v", lim.Path, err)
		}
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"response":"limit_reached","message":"Too many requests. Try again later."}`))
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
