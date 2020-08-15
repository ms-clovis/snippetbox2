package validation

import (
	"github.com/ms-clovis/snippetbox2/pkg/models"
	"testing"
)

//func TestIsDate(t *testing.T) {
//	sDate := "Mon, 01-10-1959"
//	if !IsDate(sDate){
//		t.Error("Should be date")
//	}
//	sDate = "1959-33-33"
//	if IsDate(sDate){
//		t.Error("Should not be date")
//	}
//}

func TestIsOneOfValue(t *testing.T) {
	slc := []string{"1", "2", "3"}

	if !IsOneOfValue("1", slc) {
		t.Error("1 should be member of list 1,2,3")
	}

	if IsOneOfValue("4", slc) {
		t.Error("4 should not be member of 1,2,3")
	}
}

func TestDoesStartWith(t *testing.T) {
	testStr := "foobar"
	if !DoesStartWith(testStr, "foo") {
		t.Fatal("Foobar begins with foo")
	}
	if DoesStartWith(testStr, "bar") {
		t.Fatal("Foobar does not begin with bar")
	}
}

func TestDoesEndWith(t *testing.T) {
	testStr := "foobar"
	if !DoesEndWith(testStr, "bar") {
		t.Fatal("Foobar ends with bar ")
	}
	if DoesEndWith(testStr, "foo") {
		t.Fatal("Foobar does not end with foo")
	}
}

func TestDoesContain(t *testing.T) {
	testStr := "foobar"
	if !DoesContain(testStr, "oob") {
		t.Fatal("Foobar contains 'oob'")
	}
	if DoesContain(testStr, "OOB") {
		t.Fatal("Foobar does not contain 'OOB'")
	}
}

func TestIsInteger(t *testing.T) {
	if !IsInteger("12") {
		t.Fatal("String 12 is an integer ")
	}
	if !IsInteger("0") {
		t.Fatal("String 0 is an integer ")
	}

	if !IsInteger("-1") {
		t.Fatal("String -1 is an integer ")
	}

	if !IsInteger("12234567890123456789") {
		t.Fatal("String 12234567890123456789 is a big integer")
	}
	if IsInteger("1.23") {
		t.Fatal("1.23 is not an integer")
	}
	if IsInteger("a") {
		t.Fatal("A is not an integer")
	}

}

func TestIsBlank(t *testing.T) {
	str := " x a "
	if IsBlank(str) {
		t.Error("String is not blank")
	}

	str = " \t"
	if !IsBlank(str) {
		t.Error("String is blank")
	}
}

func TestIsAuthenticated(t *testing.T) {
	password := "123456"
	alphaPW := "abcdef"
	user := &models.User{
		ID:       1,
		Name:     "foo@test.com",
		Password: password,
		Active:   true,
	}
	//pw ,_ := bcrypt.GenerateFromPassword([]byte(user.Password),bcrypt.DefaultCost)
	user.SetEncryptedPassword(password)

	isAuth := IsAuthenticated(user.Password, password)

	if !isAuth {
		t.Error("Did not authenticate")
	}
	user.SetEncryptedPassword(alphaPW)
	isAuth = IsAuthenticated(user.Password, password)

	if isAuth {
		t.Error("Did not authenticate")
	}
}

func TestIsLessThanChars(t *testing.T) {
	str := "12345"
	if !IsLessThanChars(str, 6) {
		t.Error("String is less than max")
	}

	if IsLessThanChars(str, 4) {
		t.Error("String is not less than max")
	}
}

func TestIsValidEmailAddr(t *testing.T) {
	addr := "ms.clovis@verizon.net"
	if !IsValidEmailAddr(addr) {
		t.Error("Is a correct address")
	}

	addr = "m@verizon.net"
	if !IsValidEmailAddr(addr) {
		t.Error("Is a correct address")
	}

	addr = "m@verizon"
	if IsValidEmailAddr(addr) {
		t.Error("Is not a valid address")
	}
}

func TestIsChecked(t *testing.T) {
	str := "Checked"
	if !IsChecked(str) {
		t.Error("Is checked")
	}

	str = "check"
	if IsChecked(str) {
		t.Error("Is not checked")
	}
}
