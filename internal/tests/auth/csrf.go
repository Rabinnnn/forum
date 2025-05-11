package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type csrfToken struct {
	token   string
	expires time.Time
}

var (
	tokens = make(map[string]csrfToken)
	mutex  sync.RWMutex
)

func Generate() string {
	// Generate random bytes
	b := make([]byte, 32)
	rand.Read(b)
	token := base64.URLEncoding.EncodeToString(b)

	// Store token with expiration
	mutex.Lock()
	tokens[token] = csrfToken{
		token:   token,
		expires: time.Now().Add(1 * time.Hour),
	}
	mutex.Unlock()

	return token
}

func Verify(token string) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	if storedToken, exists := tokens[token]; exists {
		if time.Now().Before(storedToken.expires) {
			// Clean up used token
			delete(tokens, token)
			return true
		}
	}
	return false
}

// Periodically clean up expired tokens
func init() {
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			mutex.Lock()
			for token, t := range tokens {
				if time.Now().After(t.expires) {
					delete(tokens, token)
				}
			}
			mutex.Unlock()
		}
	}()
}
