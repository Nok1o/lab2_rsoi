package validation

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var InvalidUsernameErr = errors.New("invalid username")

const (
	minimalUsernameLength = 3
	maximalUsernameLength = 80
)

// ValidateUsername checks a username before it is used as a database key.
func ValidateUsername(username string) error {
	if !utf8.ValidString(username) || strings.TrimSpace(username) != username {
		return InvalidUsernameErr
	}

	length := utf8.RuneCountInString(username)
	if length < minimalUsernameLength || length > maximalUsernameLength {
		return InvalidUsernameErr
	}

	return nil
}
