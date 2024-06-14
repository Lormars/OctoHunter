package checker

import (
	"testing"
)

func TestFindImmediateCname(t *testing.T) {
	res, err := FindImmediateCNAME("dlab01-mda-w2b-filemanagement.azurewebsites.net")
	t.Logf("res is %s, err is %v\n", res, err)
}
