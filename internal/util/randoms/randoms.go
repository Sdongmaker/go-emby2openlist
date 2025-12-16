package randoms

import (
	"math/rand"
	"strings"
)

// hexs 16 进制字符
var hexs = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}

// alphanums 字母数字字符（用于腾讯云等 CDN 的短随机数）
var alphanums = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
	"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
	"u", "v", "w", "x", "y", "z", "A", "B", "C", "D",
	"E", "F", "G", "H", "I", "J", "K", "L", "M", "N",
	"O", "P", "Q", "R", "S", "T", "U", "V", "W", "X",
	"Y", "Z",
}

// RandomHex 随机返回一串 16 进制的字符串, 可通过 n 指定长度
func RandomHex(n int) string {
	if n <= 0 {
		return ""
	}
	sb := strings.Builder{}
	for n > 0 {
		idx := rand.Intn(len(hexs))
		sb.WriteString(hexs[idx])
		n--
	}
	return sb.String()
}

// RandomAlphaNum 随机返回一串字母数字字符串（用于腾讯云 CDN Type-A 鉴权）
// 可通过 n 指定长度，建议使用 6 位
func RandomAlphaNum(n int) string {
	if n <= 0 {
		return ""
	}
	sb := strings.Builder{}
	for n > 0 {
		idx := rand.Intn(len(alphanums))
		sb.WriteString(alphanums[idx])
		n--
	}
	return sb.String()
}
