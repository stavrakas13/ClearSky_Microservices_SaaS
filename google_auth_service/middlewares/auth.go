package middlewares

import (
	"google_auth_service/utils"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Unauthorized - no token", http.StatusUnauthorized)
			return
		}

		_, err = utils.VerifyJWT(cookie.Value)
		if err != nil {
			http.Error(w, "Unauthorized - invalid token", http.StatusUnauthorized)
			return
		}
		// You can set claims in context if needed
		// ctx := context.WithValue(r.Context(), "claims", claims)
		// r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
