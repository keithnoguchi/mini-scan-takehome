package processing

import (
	"context"

	_ "github.com/gocql/gocql"
)

type scyllaProcessor struct{}

func (p *scyllaProcessor) Process(ctx context.Context, msg Scan) error {
	return nil
}
