package amplitude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"context"
)

const (
	DefaultQueueSize = 250
	ApiEndpoint      = "https://api.amplitude.com/2/httpapi"

	// TODO look into using the Identify endpoint
	// https://github.com/msingleton/amplitude-go/blob/master/client.go#L60
	// IdentifyEndpoint = "https://api.amplitude.com/identify"
)

// There is a character limit of 1024 characters for all string values
// https://help.amplitude.com/hc/en-us/articles/204771828#string-character-limit

type Event struct {
	UserID             string                 `json:"user_id,omitempty"`
	DeviceID           string                 `json:"device_id,omitempty"`
	EventType          string                 `json:"event_type,omitempty"`
	Time               time.Time              `json:"-"`
	TimeInMillis       int64                  `json:"timestamp,omitempty"`
	EventProperties    map[string]interface{} `json:"event_properties,omitempty"`
	UserProperties     map[string]interface{} `json:"user_properties,omitempty"`
	AppVersion         string                 `json:"app_version,omitempty"`
	Platform           string                 `json:"platform,omitempty"`
	OSName             string                 `json:"os_name,omitempty"`
	OSVersion          string                 `json:"os_version,omitempty"`
	DeviceBrand        string                 `json:"device_brand,omitempty"`
	DeviceManufacturer string                 `json:"device_manufacturer,omitempty"`
	DeviceModel        string                 `json:"device_model,omitempty"`
	DeviceType         string                 `json:"device_type,omitempty"`
	Carrier            string                 `json:"carrier,omitempty"`
	Country            string                 `json:"country,omitempty"`
	Region             string                 `json:"region,omitempty"`
	City               string                 `json:"city,omitempty"`
	DMA                string                 `json:"dma,omitempty"`
	Language           string                 `json:"language,omitempty"`
	Revenue            float64                `json:"revenu,omitempty"`
	Lat                float64                `json:"lat,omitempty"`
	Lon                float64                `json:"lon,omitempty"`
	IP                 string                 `json:"ip,omitempty"`
	IDFA               string                 `json:"idfa,omitempty"`
	ADID               string                 `json:"adid,omitempty"`
}

type Client struct {
	cancel        func()
	ctx           context.Context
	apiKey        string
	ch            chan Event
	flush         chan chan struct{}
	queueSize     int
	interval      time.Duration
	onPublishFunc func(*http.Response, error)
	httpClient    *http.Client
	Timeout       time.Duration
	apiEndpoint   string
}

func New(apiKey string, options ...Option) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		cancel:      cancel,
		ctx:         ctx,
		apiKey:      apiKey,
		Timeout:     time.Second * 10,
		ch:          make(chan Event, DefaultQueueSize),
		flush:       make(chan chan struct{}),
		queueSize:   DefaultQueueSize,
		interval:    time.Second * 15,
		httpClient:  http.DefaultClient,
		apiEndpoint: ApiEndpoint,
	}

	for _, opt := range options {
		opt(client)
	}

	go client.start()

	return client
}

func (c *Client) Publish(e Event) error {
	if !e.Time.IsZero() {
		e.TimeInMillis = e.Time.UnixNano() / int64(time.Millisecond)
	}

	select {
	case c.ch <- e:
		return nil
	default:
		return fmt.Errorf("Amplitude: Unable to send event, queue is full")
	}
}

// Event requires UserID and EventType to be set
func (c *Client) Event(e Event) error {
	if e.UserID == "" {
		return fmt.Errorf("Amplitude: missing required parameter: UserID")
	}
	if e.EventType == "" {
		return fmt.Errorf("Amplitude: missing required parameter: EventType")
	}

	return c.Publish(e)
}

func (c *Client) start() {
	timer := time.NewTimer(c.interval)

	bufferSize := 256
	buffer := make([]Event, bufferSize)
	index := 0

	for {
		timer.Reset(c.interval)

		select {
		case <-c.ctx.Done():
			return

		case <-timer.C:
			if index > 0 {
				c.publish(buffer[0:index])
				index = 0
			}

		case v := <-c.ch:
			buffer[index] = v
			index++
			if index == bufferSize {
				c.publish(buffer[0:index])
				index = 0
			}

		case v := <-c.flush:
			if index > 0 {
				c.publish(buffer[0:index])
				index = 0
			}
			v <- struct{}{}
		}
	}
}

func (c *Client) publish(events []Event) {
	body := struct {
		Events []Event `json:"events"`
		APIKey string  `json:"api_key"`
	}{
		events,
		c.apiKey,
	}
	data, err := json.Marshal(body)
	if err != nil {
		c.onPublishFunc(&http.Response{}, err)
	}

	ctx, cancel := context.WithTimeout(c.ctx, c.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiEndpoint, bytes.NewBuffer(data))
	if err != nil {
		c.onPublishFunc(&http.Response{}, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	resp, err := c.httpClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	} else {
		resp = &http.Response{}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		err = fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	c.onPublishFunc(resp, err)
}

func (c *Client) Flush() {
	ch := make(chan struct{})
	defer close(ch)

	c.flush <- ch
	<-ch
}

func (c *Client) Close() {
	c.Flush()
	c.cancel()
}
