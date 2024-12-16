package processing

import (
	"context"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
	"github.com/scylladb/gocqlx/v3/table"
)

type scyllaProcessor struct {
	// Scylla session.
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
	SortKey: []string{"port", "service"},
})

func newScyllaProcessor(cfg ProcessorConfig) (Processor, error) {
	clusterCfg := gocql.NewCluster(cfg.BackendURL().Host)
	session, err := gocqlx.WrapSession(clusterCfg.CreateSession())
	if err != nil {
		return nil, err
	}
	return &scyllaProcessor{
		session: session,
	}, nil
}

func (p *scyllaProcessor) Process(ctx context.Context, msg Scan) error {
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
	// We don't need to serialize here, as gocql.Session.Close()
	// is concurrency safe.
	p.session.Close()
}
