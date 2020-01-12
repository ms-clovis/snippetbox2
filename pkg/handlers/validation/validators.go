package validation

import (
	"golang.org/x/crypto/bcrypt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func IsAuthenticated(hashedPW string, pw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPW), []byte(pw))
	return err == nil

}

func IsLessThanChars(s string, numberOfChars int) bool {
	return utf8.RuneCountInString(s) < numberOfChars
}
func IsMoreThanChars(s string, maxChars int) bool {
	return utf8.RuneCountInString(s) > maxChars
}

func DoesStartWith(s string, beginning string) bool {
	return strings.HasPrefix(s, beginning)
}

func DoesEndWith(s string, end string) bool {
	return strings.HasSuffix(s, end)
}

func DoesContain(s string, part string) bool {
	return strings.Contains(s, part)
}

func IsValidEmailAddr(mail string) bool {
	var rx = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return len(mail) <= 254 && rx.MatchString(mail)
}

func IsInteger(s string) bool {
	_, err := strconv.Atoi(s)
	var ok bool
	if err != nil && strings.Contains(err.Error(), "value out of range") {
		n := new(big.Int)
		_, ok = n.SetString(s, 10)
	}
	return err == nil || ok
}

//func IsFloat(s string) bool {
//	f, err := strconv.ParseFloat(s, 64)
//
//	return err == nil && s == sameStr
//}

//func IsDate(s string,layout string)bool{
//
//	d,err := time.Parse(layout,s)
//
//	return err != nil && d.Year() >1
//}

func IsOneOfValue(s string, vals []string) bool {
	var set map[string]bool
	set = make(map[string]bool)
	for _, v := range vals {
		set[v] = true
	}
	return set[s]
}

func IsChecked(s string) bool {
	return strings.ToLower(s) == "checked"
}
