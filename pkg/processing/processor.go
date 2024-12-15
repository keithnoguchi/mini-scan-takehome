package processing

import (
	"context"
	"log"
)

func NewProcessor() Processor {
	// XXX returns the simple log processor for now.
	// XXX this should be controlled through the ProcessorConfig
	// XXX as a follow up.
	return &logProcessor{}
}

// Utility function to retrieve the log.Logger from the context.Context.
func logger(ctx context.Context) *log.Logger {
	if logger, ok := ctx.Value("logger").(*log.Logger); ok {
		return logger
	} else {
		return log.Default()
	}
}
