package salesforce

import (
	"bufio"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestFingerprint(t *testing.T) {
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Println("No .env file found")
	}
	file, err := os.Open("../../../list/salesforceFile")
	if err != nil {
		t.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		domain := scanner.Text()
		SalesforceScan(domain)
		//t.Logf("res is %v\n", res)

	}
}
