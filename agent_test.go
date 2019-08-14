package amplitude

import (
	"net/http"
	"os"
	"testing"
)

func TestCompiles(t *testing.T) {

	event := Event{
		UserID:    "123",
		EventType: "testing",
	}

	apiKey := os.Getenv("AMPLITUDE_API_KEY")
	client := New(apiKey, OnPublishFunc(func(resp *http.Response, err error) {
		// fmt.Println(resp.StatusCode, err)
		// io.Copy(os.Stdout, resp.Body)
	}))
	err := client.Publish(event)

	if err != nil {
		t.Error(err)
	}

	client.Close()
}
