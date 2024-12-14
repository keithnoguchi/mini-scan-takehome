package processing

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"time"
	"unicode/utf8"

	"github.com/censys/scan-takehome/pkg/scanning"
)

// Scan is the validated version of the scanning.Scan.
//
// All the fieds of the scanning.Scan are validated when it's
// unmarshaled by BinaryUnmarshaler.UnmarshalBinary interface function.
type Scan struct {
	// IP address.
	Ip net.IP

	// Port.
	Port uint16

	// Service name.
	Service string

	// Scaning timestamp.
	Timestamp time.Time

	// Service data.
	Data string

	// raw scanned data.
	raw scanning.Scan
}

// String represents the Scan data.
func (s Scan) String() string {
	return fmt.Sprintf(
		"[%s] %s:%d/%s: %s",
		s.Timestamp.UTC().Format("01/02 03:04:05"),
		s.Ip, s.Port, s.Service, s.Data,
	)
}

// BinaryUnmarshaler unmarshals a binary representation of itself.
func (s *Scan) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &s.raw); err != nil {
		return errors.Join(ErrScanData, err)
	}
	if err := s.validateIP(); err != nil {
		return errors.Join(ErrScanData, err)
	}
	if err := s.validatePort(); err != nil {
		return errors.Join(ErrScanData, err)
	}
	if err := s.validateData(); err != nil {
		return errors.Join(ErrScanData, err)
	}
	s.Service = s.raw.Service
	s.Timestamp = time.Unix(s.raw.Timestamp, 0)
	return nil
}

// validateIP validates the IP address field of the scan message.
func (s *Scan) validateIP() error {
	s.Ip = net.ParseIP(s.raw.Ip)
	if s.Ip == nil {
		return ErrScanIP
	}
	return nil
}

// validatePort validates the port field of the scan message.
func (s *Scan) validatePort() error {
	if s.raw.Port > math.MaxUint16 {
		return ErrScanPort
	}
	s.Port = uint16(s.raw.Port)
	return nil
}

// validateData validates the data version and the data field of
// the scan message.
func (s *Scan) validateData() error {
	hash, ok := s.raw.Data.(map[string]any)
	if !ok {
		return ErrScanDataType
	}
	switch s.raw.DataVersion {
	case scanning.V1:
		value, ok := hash["response_bytes_utf8"]
		if !ok {
			return ErrScanDataType
		}
		encoded, ok := value.(string)
		if !ok {
			return ErrScanDataType
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return errors.Join(ErrScanDataEncoding, err)
		}
		if ok := utf8.Valid(decoded); !ok {
			return ErrScanDataEncoding
		}
		s.Data = string(decoded)
	case scanning.V2:
		value, ok := hash["response_str"]
		if !ok {
			return ErrScanDataType
		}
		data, ok := value.(string)
		if !ok {
			return ErrScanDataType
		}
		s.Data = data
	default:
		return ErrScanDataVersion
	}
	return nil
}
