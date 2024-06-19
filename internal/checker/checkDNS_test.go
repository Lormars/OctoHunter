package checker

import (
	"testing"
)

func TestFindImmediateCname(t *testing.T) {
	res, cname, err := HasCname("google.com")
	t.Logf("res is %v, cname is %s, err is %v\n", res, cname, err)
}
