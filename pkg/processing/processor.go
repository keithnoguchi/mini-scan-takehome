package processing

import (
	"context"
	"log"
)

// NewProcessor returns the new Processor instance.
//
// Please take a look at ProcessorConfig type how to configure
// Processor.
func NewProcessor(cfg ProcessorConfig) Processor {
	switch cfg.BackendType() {
	case BackendScylla:
		return newScyllaProcessor(cfg)
	default:
		return &logProcessor{}
	}
}

// Utility function to retrieve the log.Logger from the context.Context.
func logger(ctx context.Context) *log.Logger {
	if logger, ok := ctx.Value("logger").(*log.Logger); ok {
		return logger
	} else {
		return log.Default()
	}
}
