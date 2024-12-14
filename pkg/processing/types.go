// Package processing provides the Internet scan message processor.
package processing

import (
	"context"

	"github.com/censys/scan-takehome/pkg/scanning"
)

// Processor interface to represent the message processor.
type Processor interface {
	Process(context.Context, scanning.Scan)
}
