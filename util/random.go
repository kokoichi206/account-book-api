package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// 全てのアルファベットを含む string。
const alphabet = "abcdefghijklmnopqrstuvwxyz"

// 初期化時に seed を固定する。
func init() {
	rand.Seed(time.Now().UnixNano())
}

// min 以上 max 以下のランダムな整数値を取得する。
// min が max より大きい時 panic を起こす。
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// n 文字のランダムな文字列を取得する。
// 文字列は全てアルファベットのみで構成される。
func RandomString(n int) string {
	var s strings.Builder
	l := len(alphabet)
	for i := 0; i < n; i++ {
		s.WriteByte(alphabet[rand.Intn(l)])
	}
	return s.String()
}

// ランダムなユーザー名を取得する。
// 7文字からなるランダムな文字列を返す。
func RandomUserName() string {
	return RandomString(7)
}

// ランダムなパスワードを取得する。
// 12文字からなるランダムな文字列を返す。
func RandomPassword() string {
	return RandomString(12)
}

// ランダムなメールアドレスを取得する。
// メールアドレスの形式に則った文字列を返す。
func RandomEmail() string {
	return fmt.Sprintf("%s@example.com", RandomString(6))
}

// ランダムな年齢を取得する。
// ２桁のランダムな数値を返す。
func RandomAge() int32 {
	return int32(RandomInt(10, 99))
}

// ランダムな残高を取得する。
// １億までのランダムな数値を返す。
func RandomBalance() int64 {
	return RandomInt(0, 1_0000_0000)
}
