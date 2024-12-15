package processing

import (
	"context"
	"log"
)

type logger struct{}

func (p *logger) Process(ctx context.Context, msg Scan) error {
	log.Printf("%s", msg)
	return nil
}
