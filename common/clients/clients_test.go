package clients

import (
	"fmt"
	"testing"

	"github.com/joho/godotenv"
)

func TestClients(t *testing.T) {
	godotenv.Load("../../.env")
	resp, err := NormalClient.Get("https://www.hackerone.com/resources/penetration-tests#top%22")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)
}
