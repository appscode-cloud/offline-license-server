package emailproviders

import (
	"strings"
)

var empty = struct{}{}

func Domain(str string) string {
	idx := strings.LastIndexByte(str, '@')
	if idx == -1 {
		return str
	}
	return str[idx+1:]
}

func IsPublicEmail(email string) bool {
	_, found := publicEmailServices[Domain(email)]
	return found
}

func IsDisposableEmail(email string) bool {
	_, found := disposableEmailServices[Domain(email)]
	return found
}

func IsWorkEmail(email string) bool {
	domain := Domain(email)

	if _, found := publicEmailServices[domain]; found {
		return false
	}
	if _, found := disposableEmailServices[domain]; found {
		return false
	}
	return true
}
