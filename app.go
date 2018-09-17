package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Effyis/stream-client-go/stream"
)

var username, password, dataSource, streamName, subscriptionName, customerName, domain string

func main() {
	flag.StringVar(&username, "U", "", "Client username (required)")
	flag.StringVar(&password, "W", "", "Client password (required)")
	flag.StringVar(&dataSource, "ds", "", "Data source (required)")
	flag.StringVar(&streamName, "sn", "", "Stream name (required)")
	flag.StringVar(&subscriptionName, "sb", "", "Subscription name (required)")
	flag.StringVar(&customerName, "cn", "", "Customer name (required)")
	flag.StringVar(&domain, "domain", "socialgist.com", "Endpoint domain")
	flag.Parse()
	if username == "" || password == "" || dataSource == "" || streamName == "" || subscriptionName == "" || customerName == "" {
		fmt.Println("Invalid argument(s), please use:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	sc := stream.NewClient(stream.Connection{
		Domain:           domain,
		Username:         username,
		Password:         password,
		DataSource:       dataSource,
		StreamName:       streamName,
		SubscriptionName: subscriptionName,
		CustomerName:     customerName,
	})
	sc.Start()
	for {
		select {
		case msg := <-sc.ChanMessage:
			fmt.Println(msg)
		case err := <-sc.ChanError:
			fmt.Fprintln(os.Stderr, err)
		case <-sc.ChanStop:
			os.Exit(0)
		}
	}
}
