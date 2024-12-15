package processing

import "context"

type logProcessor struct{}

func (p *logProcessor) Process(ctx context.Context, msg Scan) error {
	logger(ctx).Printf("%s", msg)
	return nil
}
