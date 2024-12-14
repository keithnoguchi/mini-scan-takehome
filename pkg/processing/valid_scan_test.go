package processing

import (
	"testing"

	"github.com/censys/scan-takehome/pkg/scanning"
)

func TestValidScanDataString(t *testing.T) {
	tests := []struct {
		name string
		scan ValidScan
		want error
	}{
		{
			name: "Valid v1 data",
			scan: ValidScan{
				Scan: scanning.Scan{
					DataVersion: scanning.V1,
					Data: &scanning.V1Data{
						ResponseBytesUtf8: []byte("some data"),
					},
				},
			},
			want: nil,
		},
		{
			name: "Invalid v1 data type",
			scan: ValidScan{
				Scan: scanning.Scan{
					DataVersion: scanning.V1,
					Data: &scanning.V2Data{
						ResponseStr: "some data",
					},
				},
			},
			want: ErrDataType,
		},
		{
			name: "Invalid v1 data encoding",
			scan: ValidScan{
				Scan: scanning.Scan{
					DataVersion: scanning.V1,
					Data: &scanning.V1Data{
						ResponseBytesUtf8: []byte{0x80},
					},
				},
			},
			want: ErrDataEncoding,
		},
		{
			name: "Valid v2 data",
			scan: ValidScan{
				Scan: scanning.Scan{
					DataVersion: scanning.V2,
					Data: &scanning.V2Data{
						ResponseStr: "some data",
					},
				},
			},
			want: nil,
		},
		{
			name: "Invalid v2 data type",
			scan: ValidScan{
				Scan: scanning.Scan{
					DataVersion: scanning.V2,
					Data: &scanning.V1Data{
						ResponseBytesUtf8: []byte("some data"),
					},
				},
			},
			want: ErrDataType,
		},
		{
			name: "Unsupported data version",
			scan: ValidScan{
				Scan: scanning.Scan{
					DataVersion: 0,
				},
			},
			want: ErrUnsupportedDataVersion,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.scan.DataString()
			if got, want := err, tt.want; got != want {
				t.Errorf("\ngot:  %v\nwant: %v", got, want)
			}
		})
	}
}
