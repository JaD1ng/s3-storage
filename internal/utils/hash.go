package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// CalculateMD5 计算数据的MD5哈希
func CalculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}