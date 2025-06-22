package middlewares

import (
	"context"
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

		claims, err := utils.VerifyJWT(cookie.Value)
		if err != nil {
			http.Error(w, "Unauthorized - invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to context for handlers to use
		ctx := context.WithValue(r.Context(), "claims", claims)
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "student_id", claims.StudentID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
