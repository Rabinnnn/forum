package auth

import (
    "crypto/rand"
    "encoding/base64"
    "sync"
    "time"
)

type CsrfToken struct {
    Token   string
    Expires time.Time
}

var (
    Tokens = make(map[string]CsrfToken)
    Mutex  sync.RWMutex
)

func Generate() string {
    // Generate random bytes
    b := make([]byte, 32)
    rand.Read(b)
    token := base64.URLEncoding.EncodeToString(b)

    // Store token with expiration
    Mutex.Lock()
    Tokens[token] = CsrfToken{
        Token:   token,
        Expires: time.Now().Add(1 * time.Hour),
    }
    Mutex.Unlock()

    return token
}

func Verify(token string) bool {
    Mutex.RLock()
    defer Mutex.RUnlock()

    if storedToken, exists := Tokens[token]; exists {
        if time.Now().Before(storedToken.Expires) {
            // Clean up used token
            delete(Tokens, token)
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
            Mutex.Lock()
            for token, t := range Tokens {
                if time.Now().After(t.Expires) {
                    delete(Tokens, token)
                }
            }
            Mutex.Unlock()
        }
    }()
}