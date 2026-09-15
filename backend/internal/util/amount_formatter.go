package util

import (
	"strconv"
	"strings"
)

// FormatAmount 格式化金额数字为千分位字符串。
func FormatAmount(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	sign := ""
	if strings.HasPrefix(intPart, "-") {
		sign = "-"
		intPart = strings.TrimPrefix(intPart, "-")
	}
	var b strings.Builder
	n := len(intPart)
	for i, ch := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(ch)
	}
	return sign + b.String() + "." + parts[1]
}
