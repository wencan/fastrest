package restutils

import "strings"

// CompareHumanizeString 比较两个字符串，结果接近可视结果。
func CompareHumanizeString(a, b string) int {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	return strings.Compare(a, b)
}
