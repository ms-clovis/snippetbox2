package validation

import "testing"

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

func TestIsLessThanMaxChars(t *testing.T) {
	str := "12345"
	if !IsLessThanMaxChars(str, 6) {
		t.Error("String is less than max")
	}

	if IsLessThanMaxChars(str, 4) {
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
