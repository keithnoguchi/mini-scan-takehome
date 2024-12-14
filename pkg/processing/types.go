// Package processing provides the Internet scan message processor.
package processing

import (
	"context"
	"errors"

	"github.com/censys/scan-takehome/pkg/scanning"
)

// ErrScanInvalid indicates the invalid scan data.
var ErrScanInvalid = errors.New("Invalid scan data")

// ErrDataVersion indicates the unsupported DataVersion.
var ErrUnsupportedDataVersion = errors.New("Wrong data version")

// ErrDataEncoding indicates the wrong data type for the data version.
var ErrDataType = errors.New("Wrong data encoding")

// ErrDataEncoding indicates the wrong data encoding.
var ErrDataEncoding = errors.New("Wrong data encoding")

// Processor interface to represent the message processor.
type Processor interface {
	Process(context.Context, scanning.Scan)
}
