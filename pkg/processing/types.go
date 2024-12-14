// Package processing provides the Internet scan message processor.
package processing

import (
	"context"
	"errors"
)

// ErrScanData indicates the invalid scan data.
//
// This is the top node of the scan data related errors.  Use errors.Is
// to see if the error is related to the scan data or not.
//
//	var scan processing.Scan
//	if err := scan.UnmarshalBinary(msg.Data); err != nil {
//	    if errors.Is(err, processing.ErrScanData) {
//	      // It's the invalid scan data.
//	    }
//	}
var ErrScanData = errors.New("Invalid scan data")

// ErrScanIP indicates the invalid IP address scan data.
var ErrScanIP = errors.New("Invalid scan IP address")

// ErrScanPort indicates the invalid port scan data.
var ErrScanPort = errors.New("Invalid scan port")

// ErrScanDataVersion indicates the unsupported DataVersion.
var ErrScanDataVersion = errors.New("Unsupported scan data version")

// ErrScanDataType indicates the wrong data type for the data version.
var ErrScanDataType = errors.New("Invalid scan data type")

// ErrScanDataEncoding indicates the wrong scan data encoding.
var ErrScanDataEncoding = errors.New("Invalid scan data encoding")

// Processor interface to represent the message processor.
type Processor interface {
	Process(context.Context, Scan) error
}
