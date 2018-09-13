// Package stream provides access to stream client.
package stream

import (
	"bufio"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const reconnectDelay = 60 * time.Second

// Connection is used to configure stream client connection configuration.
type Connection struct {
	Domain       string
	Username     string
	Password     string
	DataSource   string
	StreamName   string
	CustomerName string
}

// Client is a stream client that is able to consume single stream API endpoint with auto-reconnect built in.
type Client struct {
	Connection Connection
	// Channel for receiving messages.
	// Start() must be called first.
	ChanMessage chan string
	// Channel for receiving stop signal.
	// It will emit after Stop() is called.
	ChanStop chan bool
	// Channel for receiving errors.
	// It will emit when ever error occurred.
	// These errors are not fatal becaue client will automatically reconnect but channel still has to be consumed.
	ChanError chan error
	// Maximum buffer size per message. Default is 67108864 (64MB).
	// If message is bigger then MaxLineBufferSize you'll see `bufio.Scanner: token too long` error.
	MaxLineBufferSize int

	client   http.Client
	mutexRun sync.Mutex
	running  bool
	chStop   chan bool
}

// NewClient creates Client.
func NewClient(connection Connection) *Client {
	client := Client{
		Connection:        connection,
		ChanStop:          make(chan bool),
		ChanMessage:       make(chan string),
		ChanError:         make(chan error),
		MaxLineBufferSize: 64 * 1024 * 1024,
	}
	client.client.Timeout = 0
	if client.Connection.Domain == "" {
		client.Connection.Domain = "socialgist.com"
	}
	return &client
}

// Start begins consuming stream, returns false if it's already consuming.
func (sc *Client) Start() bool {
	sc.mutexRun.Lock()
	defer sc.mutexRun.Unlock()
	if !sc.running {
		sc.running = true
		sc.chStop = make(chan bool)
		go func() {
			for {
				sc.mutexRun.Lock()
				if !sc.running {
					sc.mutexRun.Unlock()
					break
				}
				sc.mutexRun.Unlock()
				err := sc.runStream()
				if err != nil {
					sc.ChanError <- err
				}
				select {
				case <-time.After(reconnectDelay):
				case <-sc.chStop:
					break
				}
			}
		}()
		return true
	}
	return false
}

// Stop ends previously started consumption, returns false if it's already stopped.
func (sc *Client) Stop() bool {
	sc.mutexRun.Lock()
	defer sc.mutexRun.Unlock()
	if sc.running {
		sc.running = false
		close(sc.chStop)
		go func() {
			sc.ChanStop <- true
		}()
		return true
	}
	return false
}

func (sc *Client) runStream() error {
	url := fmt.Sprintf("https://%s.%s/stream/%s_%s/subscription/%s/part/1/data.json",
		sc.Connection.CustomerName,
		sc.Connection.Domain,
		sc.Connection.DataSource,
		sc.Connection.StreamName,
		sc.Connection.StreamName,
	)
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(sc.Connection.Username, sc.Connection.Password)
	resp, err := sc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Invalid response status: %d", resp.StatusCode)
	}
	scanner := bufio.NewScanner(resp.Body)
	buffer := make([]byte, 64*1024)
	chDone := make(chan struct{})
	defer close(chDone)
	go func() {
		select {
		case <-sc.chStop:
		case <-chDone:
		}
		resp.Body.Close()
	}()
	scanner.Buffer(buffer, sc.MaxLineBufferSize)
	for scanner.Scan() {
		sc.ChanMessage <- scanner.Text()
	}
	return scanner.Err()
}
