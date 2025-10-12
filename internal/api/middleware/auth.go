// internal/api/middleware/auth.go
package middleware

import (
	"context"
	"fmt"
	"github.com/chestorix/gophkeeper/internal/interfaces"
	"net/http"
	"strings"
)

const UserIDKey = "userID"

func Auth(authService interfaces.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			fmt.Println("Received token:", token)

			if strings.HasPrefix(token, "Bearer ") {
				token = strings.TrimPrefix(token, "Bearer ")
			}

			userID, err := authService.ValidateToken(token)
			if err != nil {
				fmt.Println("Token validation error:", err)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			fmt.Println("Validated userID:", userID)

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
