package jwt

import (
	"crypto/sha256"
	"encoding/hex"
)

/*
* @Author: zouyx
* @Email: zouyx@knowsec.com
* @Date:   2025/11/19 上午11:55
* @Package:
 */

func Hash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
