package validationutils

import (
	"net/url"
	"regexp"
	"strconv"
	"unicode"
)

var (
	emailRegexp      = regexp.MustCompile(`^[A-Za-z0-9+_.-]+@[A-Za-z0-9.-]+$`)
	urlRegexp        = regexp.MustCompile(`^(https?://)?([\da-z.-]+)\.([a-z.]{2,6})[/\w .-]*/?$`)
	phoneRegexp      = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	numericRegexp    = regexp.MustCompile(`^\d+$`)
	alphabeticRegexp = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphaNumRegexp   = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	fileNameRegexp   = regexp.MustCompile(`^[\w\-. ]+$`)
)

// IsValidEmail validates email address format.
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	return emailRegexp.MatchString(email)
}

// IsValidUrl validates basic URL format.
func IsValidUrl(urlStr string) bool {
	if urlStr == "" {
		return false
	}
	return urlRegexp.MatchString(urlStr)
}

// IsValidUrlStrict validates URL strictly using net/url.
func IsValidUrlStrict(urlStr string) bool {
	if urlStr == "" {
		return false
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// IsValidPhoneNumber validates phone numbers using international format.
func IsValidPhoneNumber(phone string) bool {
	if phone == "" {
		return false
	}
	return phoneRegexp.MatchString(phone)
}

// IsValidCreditCard validates a credit card number using the Luhn Algorithm.
func IsValidCreditCard(cardNumber string) bool {
	if cardNumber == "" {
		return false
	}

	sum := 0
	alternate := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		n, err := strconv.Atoi(string(cardNumber[i]))
		if err != nil {
			return false
		}
		if alternate {
			n *= 2
			if n > 9 {
				n = (n % 10) + 1
			}
		}
		sum += n
		alternate = !alternate
	}
	return sum%10 == 0
}

// IsNumeric checks if the string contains only digits.
func IsNumeric(str string) bool {
	if str == "" {
		return false
	}
	return numericRegexp.MatchString(str)
}

// IsAlphabetic checks if the string contains only letters.
func IsAlphabetic(str string) bool {
	if str == "" {
		return false
	}
	return alphabeticRegexp.MatchString(str)
}

// IsAlphanumeric checks if the string contains only alphanumeric characters.
func IsAlphanumeric(str string) bool {
	if str == "" {
		return false
	}
	return alphaNumRegexp.MatchString(str)
}

// IsStrongPw checks password strength.
// - At least 8 characters
// - At least one upper case letter
// - At least one lower case letter
// - At least one digit
// - At least one special character
func IsStrongPw(pw string) bool {
	if len(pw) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, c := range pw {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		default:
			hasSpecial = true
		}

		if hasUpper && hasLower && hasDigit && hasSpecial {
			return true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// IsValidFileName checks if the string is a valid file name.
func IsValidFileName(fileName string) bool {
	if fileName == "" {
		return false
	}
	return fileNameRegexp.MatchString(fileName)
}
