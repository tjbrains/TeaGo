package rands

import "math/rand/v2"

const (
	hexChars          = "0123456789abcdef"
	hexCharsLength    = len(hexChars)
	letterChars       = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterCharsLength = len(letterChars)
)

// String 获取随机字符串
func String(n int) string {
	if n <= 0 {
		return ""
	}
	var b = make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = letterChars[rand.Int()%letterCharsLength]
	}
	return string(b)
}

// HexString 获取随机的一个16进制的字符串，即返回的字符串中只包含[0-9a-f]字符
func HexString(n int) string {
	if n <= 0 {
		return ""
	}
	var b = make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = hexChars[rand.Int()%hexCharsLength]
	}
	return string(b)
}
