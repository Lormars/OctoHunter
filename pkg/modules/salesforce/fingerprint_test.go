package salesforce

import (
	"bufio"
	"os"
	"testing"
)

func TestFingerprint(t *testing.T) {
	file, err := os.Open("../../../list/salesforceFile")
	if err != nil {
		t.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		domain := scanner.Text()
		res, _ := Fingerprint(domain)
		t.Logf("res is %v\n", res)
	}
}
