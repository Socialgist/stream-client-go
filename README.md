Stream client application and library for golang.

## Use as an app
Download latest binary from release page.

## Use as a library

### Docs
[https://godoc.org/github.com/Effyis/stream-client-go/stream](https://godoc.org/github.com/Effyis/stream-client-go/stream)

### Example
```go
package main

import (
    "fmt"
    "os"

    "github.com/Effyis/stream-client-go/stream"
)

func main() {
    // create new client instance
    sc := stream.NewClient(stream.Connection{
        Username:         "test",
        Password:         "123",
        DataSource:       "source1",
        StreamName:       "stream1",
        SubscriptionName: "sub1",
        CustomerName:     "myname",
    })
    // start the client
    sc.Start()
    // read events through channels
    for {
        select {
        case msg := <-sc.ChanMessage:
            // handle message
            fmt.Println(msg)
        // error channel must be consumed!
        case err := <-sc.ChanError:
            // got disconnected but no worries, client will automatically reconnect
            fmt.Fprintln(os.Stderr, err)
        case <-sc.ChanStop:
            // triggered by cl.Stop() call
            os.Exit(0)
        }
    }
}
```
