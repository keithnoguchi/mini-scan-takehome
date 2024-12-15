package processing

import (
	"context"

	"github.com/gocql/gocql"
)

type scyllaProcessor struct {
	cfg     *gocql.ClusterConfig
	session *gocql.Session
}

var _ Processor = &scyllaProcessor{}

func newScyllaProcessor(cfg ProcessorConfig) Processor {
	return &scyllaProcessor{
		cfg: gocql.NewCluster(cfg.BackendURL().Host),
	}
}

func (p *scyllaProcessor) Process(ctx context.Context, msg Scan) error {
	if err := p.getSession(); err != nil {
		return err
	}
	logger(ctx).Printf("%#v", p.session)
	return nil
}

func (p *scyllaProcessor) Close(ctx context.Context) {
	if p.session != nil {
		p.session.Close()
	}
}

// lazy session creation.
func (p *scyllaProcessor) getSession() (err error) {
	if p.session != nil && !p.session.Closed() {
		return nil
	}
	p.session, err = p.cfg.CreateSession()
	return err
}
