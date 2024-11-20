package utils

import "strconv"

// Unsafe conversion. Mainly used for mapping chat ids back and forth
// as discord and telebot are using strings and integres respectively.
func S2I(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
