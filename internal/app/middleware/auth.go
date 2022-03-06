package middleware

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type cipherData struct {
	key    []byte
	nonce  []byte
	aesGCM cipher.AEAD
}

var cipherInstance *cipherData

func cipherInit() error {
	if cipherInstance == nil {
		key, err := generateRandom(2 * aes.BlockSize)
		if err != nil {
			return err
		}

		aesblock, err := aes.NewCipher(key)
		if err != nil {
			return err
		}

		aesgcm, err := cipher.NewGCM(aesblock)
		if err != nil {
			return err
		}

		nonce, err := generateRandom(aesgcm.NonceSize())
		if err != nil {
			return err
		}
		cipherInstance = &cipherData{key: key, aesGCM: aesgcm, nonce: nonce}
	}
	return nil
}
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func encrypt(userID string) (string, error) {
	if err := cipherInit(); err != nil {
		return "", err
	}
	encrypted := cipherInstance.aesGCM.Seal(nil, cipherInstance.nonce, []byte(userID), nil)
	return hex.EncodeToString(encrypted), nil
}

func decrypt(token string) (string, error) {
	if err := cipherInit(); err != nil {
		return "", err
	}
	b, err := hex.DecodeString(token)
	if err != nil {
		return "", err
	}
	userID, err := cipherInstance.aesGCM.Open(nil, cipherInstance.nonce, b, nil)
	if err != nil {
		return "", err
	}
	return string(userID), nil
}

func AuthHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := uuid.NewString()
		validAccessToken := false
		if c, err := r.Cookie(AccessToken); err == nil {
			if decrypted, err := decrypt(c.Value); err == nil {
				userID = decrypted
				log.Printf("Decrypted userID: '%s'", userID)
				validAccessToken = true
			}
		}
		if !validAccessToken {
			// cookie not found or not valid
			encrypted, err := encrypt(userID)
			if err != nil {
				http.Error(w, "Can not encrypt token", http.StatusInternalServerError)
				return
			}
			log.Printf("Set cookie '%s' for current userID: '%s'", encrypted, userID)
			c := &http.Cookie{
				Name:  AccessToken,
				Value: encrypted,
				Path:  `/`,
			}
			http.SetCookie(w, c)
		}
		ctx := context.WithValue(r.Context(), UserCtx, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
