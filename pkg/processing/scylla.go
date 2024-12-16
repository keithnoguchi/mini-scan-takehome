package processing

import (
	"context"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
	"github.com/scylladb/gocqlx/v3/table"
)

type scyllaProcessor struct {
	cfg     *gocql.ClusterConfig
	session gocqlx.Session
}

var _ Processor = &scyllaProcessor{}

// serviceTable allows for the simple CRUD operation.
//
// table.Meetadata specifies table name and columns.  It must be in sync
// with the schema.
var servicesTable = table.New(table.Metadata{
	Name:    "censys.services",
	Columns: []string{"ip", "port", "service", "data", "timestamp"},
	PartKey: []string{"ip"},
	SortKey: []string{"port"},
})

func newScyllaProcessor(cfg ProcessorConfig) Processor {
	return &scyllaProcessor{
		cfg: gocql.NewCluster(cfg.BackendURL().Host),
	}
}

func (p *scyllaProcessor) Process(ctx context.Context, msg Scan) error {
	if err := p.getSession(); err != nil {
		return err
	}
	// WithTimestamp gurantees the latest scanned entries in the
	// ScyllaDB datastore.
	//
	// Please take a look at the documentation below for more detail:
	//
	// go doc github.com/scylladb/gocqlx/v3.Queryx.WithTimestamp
	return p.session.Query(servicesTable.Insert()).BindStruct(msg).
		WithTimestamp(msg.Timestamp.Unix()).
		ExecRelease()
}

func (p *scyllaProcessor) Close(ctx context.Context) {
	if p.session.Session != nil && !p.session.Closed() {
		p.session.Close()
	}
}

// lazy session creation.
func (p *scyllaProcessor) getSession() (err error) {
	if p.session.Session != nil && !p.session.Closed() {
		return nil
	}
	p.session, err = gocqlx.WrapSession(p.cfg.CreateSession())
	return err
}
