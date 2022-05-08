package util

import (
	"golang.org/x/crypto/bcrypt"
)

// bcryptを使ってハッシュ化されたパスワードを取得する。
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// ハッシュ化されたパスワードと生のパスワードが等しいものかをチェックする。
func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
