package processing

import (
	"context"
	"sync"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
	"github.com/scylladb/gocqlx/v3/table"
)

type scyllaProcessor struct {
	// Scylla cluster configuration.
	clusterCfg *gocql.ClusterConfig

	// For lazy initialization of Scylla session.
	sessionOnce sync.Once

	// Session creation error.
	sessionErr error

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

func newScyllaProcessor(cfg ProcessorConfig) Processor {
	return &scyllaProcessor{
		clusterCfg:  gocql.NewCluster(cfg.BackendURL().Host),
		sessionOnce: sync.Once{},
	}
}

func (p *scyllaProcessor) Process(ctx context.Context, msg Scan) error {
	if err := p.aquireSession(); err != nil {
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
	// We don't need the memory serialization here, as
	// gocql.Session.Close() is concurrency safe.
	p.session.Close()
}

// Utility function to acquire the ScyllaDB session lazily.
func (p *scyllaProcessor) aquireSession() error {
	p.sessionOnce.Do(func() {
		session, err := gocqlx.WrapSession(p.clusterCfg.CreateSession())
		if err != nil {
			// We record the error when we couldn't create
			// a session and keep the error for the later
			// reference.
			p.sessionErr = backendError(err)
			return
		}
		p.session = session
	})
	return p.sessionErr
}
