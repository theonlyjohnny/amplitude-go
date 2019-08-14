amplitude-go
============

Amplitude client for Go. For additional documentation visit https://amplitude.com/docs or view the godocs.


## Background

This is a fork of the original that combines [ConradIrwin's modification](https://github.com/savaki/amplitude-go/pull/1/commits/24bed477b09f634b75b3dcfc41332e63c10cd0dc) to use a custom http.Client and the better event options of https://github.com/msingleton/amplitude-go. I've also started work on actual testing.

TODO:

- More testing
- Support for the identify endpoint

## Installation

	$ go get github.com/xeoncross/amplitude-go

## Examples

### Basic Client

Full example of a simple event tracker.

```go
	apiKey := os.Getenv("AMPLITUDE_API_KEY")
	client := amplitude.New(apiKey)
	client.Publish(amplitude.Event{
		UserId:    "123",
		EventType: "sample",
	})
```
