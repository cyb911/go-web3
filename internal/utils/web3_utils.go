package utils

import "strings"

func ParseEthToWei(amount string) string {
	parts := strings.Split(amount, ".")
	if len(parts) == 1 {
		return parts[0] + "000000000000000000"
	}

	decimals := parts[1]
	if len(decimals) > 18 {
		decimals = decimals[:18] // 截断
	}
	for len(decimals) < 18 {
		decimals += "0"
	}
	return parts[0] + decimals
}
