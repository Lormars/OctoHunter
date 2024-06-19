package salesforce

import "testing"

func TestFingerprint(t *testing.T) {
	res, _ := Fingerprint("https://example.com/")
	t.Logf("res is %v\n", res)
}
