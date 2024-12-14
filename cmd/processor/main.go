// A mini-scan-takehome scan processor.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/pubsub"
	"github.com/censys/scan-takehome/pkg/processing"
)

func main() {
	// Parses the command line arguments.
	projectId := flag.String(
		"project-id",
		"test-project",
		"GCP Project ID",
	)
	subscriptionId := flag.String(
		"subscription-id",
		"scan-sub",
		"GCP subscription ID",
	)
	flag.Parse()

	// Creates a pubsub client and subscribes to the pre-existing
	// subscription channel.
	ctx, cancel := context.WithCancel(context.Background())
	client, err := pubsub.NewClient(ctx, *projectId)
	if err != nil {
		log.Fatal(err)
	}
	sub := client.Subscription(*subscriptionId)
	log.Printf("Subscribed to %s\n", sub)

	// Create a processor to process messages.
	processor := processing.NewProcessor()
	receiver := func(ctx context.Context, msg *pubsub.Message) {
		var scan processing.ValidScan
		if err := scan.UnmarshalBinary(msg.Data); err != nil {
			log.Printf("Dropping the invalid scan data")
			msg.Ack()
			return
		}
		processor.Process(ctx, scan.Scan)
		msg.Ack()
	}
	if err := sub.Receive(ctx, receiver); err != nil {
		log.Fatalf("Pub/Sub receive failed: %v", err)
	}

	// Cancel the context once the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Signal(%v) received", sig)
	cancel()

	// Wait for the child goroutines to cancel.
	<-ctx.Done()
	err = ctx.Err()
	switch err {
	case nil, context.Canceled:
		log.Println("Gracefully shutdown the process")
	default:
		log.Fatalf(
			"Received error(%v) during the shutdown process",
			err,
		)
	}
}
