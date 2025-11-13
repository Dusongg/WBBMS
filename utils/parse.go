package utils

import (
	"strconv"
	"strings"
)

// ParseUintSlice 解析逗号分隔的uint字符串为切片
func ParseUintSlice(s string) ([]uint, error) {
	if s == "" {
		return []uint{}, nil
	}

	parts := strings.Split(s, ",")
	result := make([]uint, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		num, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, err
		}
		result = append(result, uint(num))
	}

	return result, nil
}

