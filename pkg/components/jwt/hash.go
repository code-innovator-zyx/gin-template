package jwt

import (
	"crypto/sha256"
	"encoding/hex"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/19 上午11:55
* @Package:
 */

func Hash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// SecureCompare constant-time comparison
func SecureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	// Use HMAC to compare in constant time
	ah := []byte(a)
	bh := []byte(b)
	var res byte
	for i := 0; i < len(ah); i++ {
		res |= ah[i] ^ bh[i]
	}
	return res == 0
}
