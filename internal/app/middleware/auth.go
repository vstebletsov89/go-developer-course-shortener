// Package middleware provides primitives for authorization and compress services.
package middleware

import (
	"context"
	"github.com/google/uuid"
	"go-developer-course-shortener/internal/app/service"
	"log"
	"net/http"
)

// AuthHandle implements authorization handler.
// This handler is used as a middleware for all server requests.
func AuthHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := uuid.NewString()
		validAccessToken := false
		if c, err := r.Cookie(service.AccessToken); err == nil {
			if decrypted, err := service.Decrypt(c.Value); err == nil {
				userID = decrypted
				log.Printf("Decrypted userID: '%s'", userID)
				validAccessToken = true
			}
		}
		if !validAccessToken {
			// cookie not found or not valid
			encrypted, err := service.Encrypt(userID)
			if err != nil {
				http.Error(w, "Can not encrypt token", http.StatusInternalServerError)
				return
			}
			log.Printf("Set cookie '%s' for current userID: '%s'", encrypted, userID)
			c := &http.Cookie{
				Name:  service.AccessToken,
				Value: encrypted,
				Path:  `/`,
			}
			http.SetCookie(w, c)
		}
		ctx := context.WithValue(r.Context(), service.UserCtx, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
