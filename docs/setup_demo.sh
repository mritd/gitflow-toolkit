#!/bin/bash
# Setup demo files for VHS recording

rm -rf ~/tmp/testgit
mkdir -p ~/tmp/testgit
cd ~/tmp/testgit
git init
git config user.name "Demo User"
git config user.email "demo@example.com"

# Create initial commit
echo "# Test Project" > README.md
git add README.md
git commit -m "init"

mkdir -p auth

cat > auth/handler.go << 'EOF'
package auth

import "net/http"

// LoginHandler handles user login requests
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := Authenticate(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	RevokeToken(r.Header.Get("Authorization"))
	w.WriteHeader(http.StatusNoContent)
}
EOF

cat > auth/token.go << 'EOF'
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

var tokenStore = make(map[string]time.Time)

// GenerateToken creates a new JWT token
func GenerateToken(userID string) string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)
	tokenStore[token] = time.Now().Add(24 * time.Hour)
	return token
}

// ValidateToken checks if token is valid and not expired
func ValidateToken(token string) bool {
	exp, ok := tokenStore[token]
	return ok && time.Now().Before(exp)
}

// RevokeToken invalidates a token
func RevokeToken(token string) {
	delete(tokenStore, token)
}
EOF

git add auth/
echo "Demo files ready!"
