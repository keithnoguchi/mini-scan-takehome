package processing

import "context"

type logProcessor struct{}

var _ Processor = &logProcessor{}

func (p *logProcessor) Process(ctx context.Context, msg Scan) error {
	logger(ctx).Printf("%s", msg)
	return nil
}

func (p *logProcessor) Close(ctx context.Context) {}
