package main

import (
	"fmt"
	"net/http"
	"os"

	amplitude "github.com/Xeoncross/amplitude-go"
)

func main() {
	apiKey := os.Getenv("AMPLITUDE_API_KEY")

	client := amplitude.New(apiKey, amplitude.OnPublishFunc(func(resp *http.Response, err error) {
		fmt.Fprintf(os.Stderr, "status: %v, err: %v\n", resp.StatusCode, err)
		// io.Copy(os.Stdout, resp.Body)
	}))

	client.Publish(amplitude.Event{
		UserID:    "123",
		EventType: "sample",
	})

	client.Close()
}
