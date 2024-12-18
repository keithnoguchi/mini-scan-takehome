// A mini-scan-takehome scan processor.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"

	"github.com/censys/scan-takehome/pkg/processing"
)

func main() {
	projectId := flag.String("project-id", "test-project", "GCP Project ID")
	subId := flag.String("subscription-id", "scan-sub", "Pub/Sub subscription ID")
	n := flag.Int("concurrency", 2, "Number of concurrent processors")
	backend := flag.String("backend-type", "scylla", "Processor backend")
	backendURL := flag.String("backend-url", "//scylladb:9042", "Processor backend URL")
	flag.Parse()

	// Creates the Pub/Sub processor builder.
	ctx, cancel := context.WithCancel(context.Background())
	cfg := processing.ProcessorConfig{
		"projectId":   *projectId,
		"backendType": *backend,
		"backendURL":  *backendURL,
	}
	b, err := NewBuilder(ctx, cfg)
	if err != nil {
		log.Fatalf("Can't create the processor builder: %v", err)
	}

	// Spawns the processor goroutine(s).
	var wg sync.WaitGroup
	for i := 0; i < *n; i++ {
		wg.Add(1)
		go b.build(ctx, &wg, *subId)
	}

	// Cancels the context once the signal is received.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Signal(%v) received", sig)
	cancel()

	// Wait for the processor to complete and close the processor
	// builder to clean resources.
	wg.Wait()
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	b.close(ctx)
	cancel()
	log.Println("Gracefully shutdown the processor")
}

// builder builds the new processor.
type builder struct {
	client    *pubsub.Client
	processor processing.Processor
	nextId    atomic.Uint32
}

func NewBuilder(
	ctx context.Context,
	cfg processing.ProcessorConfig,
) (*builder, error) {
	// Creates a pubsub client goroutine to process
	// the message through the subscription.
	client, err := pubsub.NewClient(ctx, cfg["projectId"].(string))
	if err != nil {
		return nil, err
	}
	processor := processing.NewProcessor(cfg)
	return &builder{
		client:    client,
		processor: processor,
	}, nil
}

func (b *builder) build(
	ctx context.Context,
	wg *sync.WaitGroup,
	subscriptionId string,
) error {
	defer wg.Done()
	name := fmt.Sprintf("[processor%02d] ", b.nextId.Add(1))
	logger := log.New(os.Stderr, name, log.LUTC)

	// Subscribe to the topic identified with the Pub/Sub subscriptionId.
	sub := b.client.Subscription(subscriptionId)
	logger.Printf("subscribed to %s\n", sub)

	// Receiver to process messages.
	receiver := func(ctx context.Context, msg *pubsub.Message) {
		ctx = context.WithValue(ctx, "logger", logger)
		var scan processing.Scan
		if err := scan.UnmarshalBinary(msg.Data); err != nil {
			logger.Printf("Dropping the invalid scan data: %v", err)
			msg.Ack()
			return
		}
		if err := b.processor.Process(ctx, scan); err != nil {
			msg.Nack()
			logger.Fatalf("Process error, exiting...: %v", err)
		} else {
			msg.Ack()
		}
	}

	// Start to process messages.
	return sub.Receive(ctx, receiver)
}

func (b *builder) close(ctx context.Context) {
	b.processor.Close(ctx)
}
