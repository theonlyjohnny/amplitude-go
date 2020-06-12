package amplitude

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestSimpleEvent(t *testing.T) {

	event := Event{
		UserID:    "123456789",
		EventType: "testing",
	}

	done := make(chan bool, 1)

	apiKey := os.Getenv("AMPLITUDE_API_KEY")
	client := New(apiKey, OnPublishFunc(func(resp *http.Response, err error) {
		if err != nil {
			io.Copy(os.Stderr, resp.Body)
			t.Error(err)
		}
		fmt.Println(resp.StatusCode, err)
		done <- true
	}))
	err := client.Publish(event)

	fmt.Println(err)
	if err != nil {
		t.Error(err)
	}

	<-done
	client.Close()
}
