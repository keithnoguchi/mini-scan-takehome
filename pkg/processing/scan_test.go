package processing

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/censys/scan-takehome/pkg/scanning"
)

func TestScanValidateIP(t *testing.T) {
	tests := []struct {
		name string
		scan Scan
		want error
	}{
		{
			name: "Valid IPv4 address",
			scan: Scan{
				raw: scanning.Scan{
					Ip: "1.1.1.1",
				},
			},
			want: nil,
		},
		{
			name: "Invalid IPv4 address",
			scan: Scan{
				raw: scanning.Scan{
					Ip: "1.1.1",
				},
			},
			want: ErrScanIP,
		},
		{
			name: "Valid IPv6 address",
			scan: Scan{
				raw: scanning.Scan{
					Ip: "2002::1",
				},
			},
			want: nil,
		},
		{
			name: "Invalid IPv6 address",
			scan: Scan{
				raw: scanning.Scan{
					Ip: "2002::0::1",
				},
			},
			want: ErrScanIP,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scan.validateIP()
			if got, want := err, tt.want; !errors.Is(got, want) {
				t.Errorf("\ngot:  %v\nwant: %v", got, want)
			}
			if tt.want == nil {
				got, want := tt.scan.Ip.String(), tt.scan.raw.Ip
				if got != want {
					t.Errorf(
						"\ngot:  %v\nwant: %v",
						got, want,
					)
				}
			}
		})
	}
}

func TestValidScanValidatePort(t *testing.T) {
	tests := []struct {
		name string
		scan Scan
		want error
	}{
		{
			name: "Valid port",
			scan: Scan{
				raw: scanning.Scan{
					Port: 1,
				},
			},
			want: nil,
		},
		{
			name: "Invalid port",
			scan: Scan{
				raw: scanning.Scan{
					Port: math.MaxUint16 + 1,
				},
			},
			want: ErrScanPort,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scan.validatePort()
			if got, want := err, tt.want; !errors.Is(got, want) {
				t.Errorf("\ngot:  %v\nwant: %v", got, want)
			}
			if tt.want == nil {
				got, want := uint32(tt.scan.Port), tt.scan.raw.Port
				if got != want {
					t.Errorf(
						"\ngot:  %v\nwant: %v",
						got, want,
					)
				}
			}
		})
	}
}

func TestValidScanValidateData(t *testing.T) {
	const msg string = "some message"
	tests := []struct {
		name string
		scan Scan
		want error
	}{
		{
			name: "Valid v1 data",
			scan: Scan{
				raw: scanning.Scan{
					DataVersion: scanning.V1,
					Data: mustData(
						scanning.V1Data{
							ResponseBytesUtf8: []byte(msg),
						},
					),
				},
			},
			want: nil,
		},
		{
			name: "Invalid v1 data type",
			scan: Scan{
				raw: scanning.Scan{
					DataVersion: scanning.V1,
					Data: mustData(
						scanning.V2Data{
							ResponseStr: "some data",
						},
					),
				},
			},
			want: ErrScanDataType,
		},
		{
			name: "Invalid v1 data encoding",
			scan: Scan{
				raw: scanning.Scan{
					DataVersion: scanning.V1,
					Data: mustData(
						scanning.V1Data{
							ResponseBytesUtf8: []byte{0x80},
						},
					),
				},
			},
			want: ErrScanDataEncoding,
		},
		{
			name: "Valid v2 data",
			scan: Scan{
				raw: scanning.Scan{
					DataVersion: scanning.V2,
					Data: mustData(
						scanning.V2Data{
							ResponseStr: msg,
						},
					),
				},
			},
			want: nil,
		},
		{
			name: "Invalid v2 data type",
			scan: Scan{
				raw: scanning.Scan{
					DataVersion: scanning.V2,
					Data: mustData(
						scanning.V1Data{
							ResponseBytesUtf8: []byte("some data"),
						},
					),
				},
			},
			want: ErrScanDataType,
		},
		{
			name: "Unsupported data version",
			scan: Scan{
				raw: scanning.Scan{
					DataVersion: 0,
					Data: mustData(
						scanning.V1Data{
							ResponseBytesUtf8: []byte("some data"),
						},
					),
				},
			},
			want: ErrScanDataVersion,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scan.validateData()
			if got, want := err, tt.want; !errors.Is(got, want) {
				t.Errorf("\ngot:  %v\nwant: %v", got, want)
			}
			if tt.want == nil {
				got, want := tt.scan.Data, msg
				if got != want {
					t.Errorf(
						"\ngot:  %v\nwant: %v",
						got, want,
					)
				}
			}
		})
	}
}

func mustData(data any) any {
	var scan scanning.Scan
	scan.Data = data
	encoded, err := json.Marshal(scan)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(encoded, &scan); err != nil {
		panic(err)
	}
	return scan.Data
}
