package processing

import (
	"encoding/json"
	"errors"
	"unicode/utf8"

	"github.com/censys/scan-takehome/pkg/scanning"
)

// ValidScan is the validated version of the scanning.Scan.
//
// All the fieds of the scanning.Scan are validated inside the
// BinaryUnmarshaler.UnmarshalBinary interface implementation.
type ValidScan struct {
	scanning.Scan
}

func (s *ValidScan) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return errors.Join(ErrScanInvalid, err)
	}
	return nil
}

// DataString returns the Data fields in the string format.
func (s *ValidScan) DataString() (string, error) {
	switch s.DataVersion {
	case scanning.V1:
		data, ok := s.Data.(*scanning.V1Data)
		if !ok {
			return "", ErrDataType
		}
		if ok := utf8.Valid(data.ResponseBytesUtf8); !ok {
			return "", ErrDataEncoding
		}
		return string(data.ResponseBytesUtf8), nil
	case scanning.V2:
		data, ok := s.Data.(*scanning.V2Data)
		if !ok {
			return "", ErrDataType
		}
		return data.ResponseStr, nil
	default:
		return "", ErrUnsupportedDataVersion
	}
}
