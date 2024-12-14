// A mini-scan-takehome scan processor.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"cloud.google.com/go/pubsub"

	"github.com/censys/scan-takehome/pkg/processing"
)

var (
	projectId = flag.String(
		"project-id",
		"test-project",
		"GCP Project ID",
	)
	subscriptionId = flag.String(
		"subscription-id",
		"scan-sub",
		"GCP subscription ID",
	)
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	flag.Parse()

	// Creates a pubsub client goroutine to process
	// the message through the subscription.
	wg.Add(1)
	go processor(ctx, &wg)

	// Cancel the context once the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Signal(%v) received", sig)
	cancel()

	// Wait for the processor to complete.
	wg.Wait()
	log.Println("Gracefully shutdown the processor")
}

// A Pub/sub scan data processor.
func processor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	client, err := pubsub.NewClient(ctx, *projectId)
	if err != nil {
		log.Fatal(err)
	}
	sub := client.Subscription(*subscriptionId)
	log.Printf("Subscribed to %s\n", sub)

	// Create a processor to process messages.
	processor := processing.NewProcessor()
	receiver := func(ctx context.Context, msg *pubsub.Message) {
		var scan processing.Scan
		if err := scan.UnmarshalBinary(msg.Data); err != nil {
			log.Printf("Dropping the invalid scan data: %v", err)
			msg.Ack()
			return
		}
		if err := processor.Process(ctx, scan); err != nil {
			msg.Nack()
			log.Fatalf("Process error, exiting...: %v", err)
		} else {
			msg.Ack()
		}
	}
	if err := sub.Receive(ctx, receiver); err != nil {
		log.Fatalf("Pub/Sub receive failed: %v", err)
	}
}
