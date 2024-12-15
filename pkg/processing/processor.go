package processing

import (
	"context"
	"log"
)

func NewProcessor(cfg ProcessorConfig) Processor {
	switch cfg.BackendType() {
	case BackendScylla:
		return &scyllaProcessor{}
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
