package processing

import (
	"context"
	"log"
)

type inMemoryProcessor struct{}

func NewProcessor() Processor {
	// XXX returns the in-memory processor for now.
	// XXX this should be controlled through the ProcessorConfig
	// XXX as a follow up.
	return &inMemoryProcessor{}
}

func (p *inMemoryProcessor) Process(ctx context.Context, msg Scan) error {
	log.Printf("%s", msg)
	return nil
}
