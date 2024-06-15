package checker

import (
	"testing"
)

func TestFindImmediateCname(t *testing.T) {
	res, err := FindImmediateCNAME("npute-prd-a-ea25687038eacb2e.elb.us-west-2.amazonaws.com")
	t.Logf("res is %s, err is %v\n", res, err)
}
