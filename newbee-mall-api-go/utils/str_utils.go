package utils

func SubStrLen(str string, length int) string {
	nameRune := []rune(str)
	if len(nameRune) > length {
		return string(nameRune[:length]) + "..."
	}
	return str
}
