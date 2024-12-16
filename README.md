# Mini-Scan

Hello!

As you've heard by now, Censys scans the internet at an incredible scale. Processing the results necessitates scaling horizontally across thousands of machines. One key aspect of our architecture is the use of distributed queues to pass data between machines.

---

The `docker-compose.yml` file sets up a toy example of a scanner. It spins up a Google Pub/Sub emulator, creates a topic and subscription, and publishes scan results to the topic. It can be run via `docker compose up`.

Your job is to build the data processing side. It should:
1. Pull scan results from the subscription `scan-sub`.
2. Maintain an up-to-date record of each unique `(ip, port, service)`. This should contain when the service was last scanned and a string containing the service's response.

> **_NOTE_**
The scanner can publish data in two formats, shown below. In both of the following examples, the service response should be stored as: `"hello world"`.
> ```javascript
> {
>   // ...
>   "data_version": 1,
>   "data": {
>     "response_bytes_utf8": "aGVsbG8gd29ybGQ="
>   }
> }
>
> {
>   // ...
>   "data_version": 2,
>   "data": {
>     "response_str": "hello world"
>   }
> }
> ```

Your processing application should be able to be scaled horizontally, but this isn't something you need to actually do. The processing application should use `at-least-once` semantics where ever applicable.

You may write this in any languages you choose, but Go, Scala, or Rust would be preferred. You may use any data store of your choosing, with `sqlite` being one example.

--- 

Please upload the code to a publicly accessible GitHub, GitLab or other public code repository account.  This README file should be updated, briefly documenting your solution. Like our own code, we expect testing instructions: whether it’s an automated test framework, or simple manual steps.

To help set expectations, we believe you should aim to take no more than 4 hours on this task.

We understand that you have other responsibilities, so if you think you’ll need more than 5 business days, just let us know when you expect to send a reply.

Please don’t hesitate to ask any follow-up questions for clarification.

---

## Solution

Here is the highlight of the processing application solution, called
[processor][]:

1. [cmd/processor/main.go][processor] as the main package, which subscribes
   to the Pub/Sub topic and spawns processor goroutines.
2. Pakcage [pkg/processing][processing] implements the processing business
   logic with the configurable backend and concurrent processing in mind.
3. Two processing backends are provided, a logger backend, for the debugging
   purpose, and a [ScyllaDB][] backend, as a persistent datastore backend
   example.
4. The [ScyllaDB][] backend demonstrates the [update ordering][update-ordering]
   through the Scylla's **USING TIMESTAMP** clause, which avoids overriding
   the scanned entries by the old/stale data.

[scylladb]: https://scylladb.com
[update-ordering]: https://opensource.docs.scylladb.com/stable/cql/dml.html#update-ordering
[processor]: ./cmd/processor/main.go
[processing]: ./pkg/processing/

### [cmd/processor][processor]

[cmd/processor][processor] is the `processor` executable main function.
It takes care of the Pub/Sub subscription operation with the processor
goroutine spawning.

Here is a snippet of the [cmd/processor][processor] main function,
which spawns multiple processor goroutines:

```golang
	// Creates the Pub/Sub processor builder.
	ctx, cancel := context.WithCancel(context.Background())
	cfg := processing.ProcessorConfig{
		"projectId":   *projectId,
		"backendType": *backend,
		"backendURL":  *backendURL,
	}
	b, err := NewBuilder(ctx, cfg)
	if err != nil {
		log.Fatalf("Can't create the Pub/Sub client: %v", err)
	}

	// Spawns the processor goroutine(s).
	var wg sync.WaitGroup
	for i := 0; i < *n; i++ {
		wg.Add(1)
		go b.build(ctx, &wg, *subId)
	}
```

It utilizes the [Builder design pattern][builder] and creates multipe
goroutines by calling `builder.build` method.

[builder]: https://en.wikipedia.org/wiki/Builder_pattern

It creates two processor goroutines to process scanned data concurrently.

You can change the number of processor goroutines through the command
line option.

Here is the currently supported processor command line options.

```bash
./processor -h
Usage of ./processor:
  -backend-type string
        Processor backend (default "scylla")
  -backend-url string
        Processor backend URL (default "//scylladb:9042")
  -concurrency int
        Number of concurrent processors (default 2)
  -project-id string
        GCP Project ID (default "test-project")
  -subscription-id string
        Pub/Sub subscription ID (default "scan-sub")
```

Here is another snippet of the [cmd/processor][processor] `builder.buid`
method.

It creates a closure, called `receiver`, and pass it to the Pub/Sub
`Subscription.Receive()` method, which blocks and process the in-coming
scanned data.

```golang
	// Receiver to process messages.
	receiver := func(ctx context.Context, msg *pubsub.Message) {
		ctx = context.WithValue(ctx, "logger", logger)
		var scan processing.Scan
		if err := scan.UnmarshalBinary(msg.Data); err != nil {
			logger.Printf("Dropping the invalid scan data: %v", err)
			msg.Ack()
			return
		}
		if err := b.processor.Process(ctx, scan); err != nil {
			msg.Nack()
			logger.Fatalf("Process error, exiting...: %v", err)
		} else {
			msg.Ack()
		}
	}

	// Start to process messages.
	return sub.Receive(ctx, receiver)
}
```

It gracefully shutdown when it receives either `SIGINT` or `SIGTERM`
signals.  It cancels the `context.Context` when it receives one of
those signals and waits for goroutines to terminate with `sync.WaitGroup`.
It finally closes the `processing.Processor` instance with the 10
seconds grace period.

```golang
	// Cancels the context once the signal is received.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Signal(%v) received", sig)
	cancel()

	// Wait for the processor to complete and close the processor
	// builder to clean resources.
	wg.Wait()
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	b.close(ctx)
	cancel()
	log.Println("Gracefully shutdown the processor")
```

### [pkg/processing][processing]

[pkg/processing][processing] implements the processing business logic.

Those are the two important data types implemented in
ithe [pkg/processing][processing] package:

1. The [processing.Scan][scan-type] type
2. The [processing.Processor][processor-interface] interface type

The [processing.Scan][scan-type] type implements the
[encoding.BinaryUnmarshaller][unmarshaller] interface, which unmarshal
and **validates** incoming binary scanned data.

The [processing.Processor][processor-interface] interface type defines
the processor backend interface, which is implemented by the logger
and the ScyllaDB processor backend.  This enables the `processor` to
support the configurable backends.

#### [processing.Scan][scan-type] type

Here is the [processing.Scan][scan-type] type:

```golang
// Scan is the validated version of the scanning.Scan.
//
// All the fieds of the scanning.Scan are validated when it's
// unmarshaled by BinaryUnmarshaler.UnmarshalBinary interface function.
type Scan struct {
        // IP address.
        Ip net.IP

        // Port.
        Port uint16

        // Service name.
        Service string

        // Scaning timestamp.
        Timestamp time.Time

        // Service data.
        Data string

        // raw scanned data.
        raw scanning.Scan
}
```

It's a Golang typed version of the [scanning.Scan][scanning-scan-type] type.

[scanning-scan-type]: ./pkg/scanning/types.go

It abstracts and centralizes the data validation logic through the
[encoding.BinaryUnarshaller][unmarshaller] implementation.  This type
is well tested by the unit tests, as it's the entry point for the
processing package.

Please take a look at the [scan_test.go][scan-test] for the unit test detail.

[scan-type]: https://github.com/keithnoguchi/mini-scan-takehome/blob/processor/pkg/processing/scan.go#L20-L38
[processor-interface]: https://github.com/keithnoguchi/mini-scan-takehome/blob/processor/pkg/processing/types.go#L39-L49
[unmarshaller]: https://pkg.go.dev/encoding#BinaryUnmarshaler
[scan-test]: ./pkg/processing/scan_test.go

#### [processing.Processor][processor-interface] interface type

The [processing.Processor][processor-interface] interface type defines
methods required to be the processor backend.

Here is the [processing.Processor][processor-interface] interface type:

```golang
// Processor interface to represent the message processor.
type Processor interface {
	// Process processes the incoming Scan data.
	Process(context.Context, Scan) error

	// Close closes the processor.
	//
	// It's a good practice to call the Close when you finish
	// using the Processor to release resources it may hold,
	// e.g. database connections.
	Close(context.Context)
}
```

It defines two methods, one for the message processing purpose and the other
for the backend cleanup purpose, e.g. closing the database connection pool.

`Process` method takes a [processing.Scan][scan-type], validated scanned data.
Since the data is already validated, the implementation of `Process` method
usually is the direct call to the backend datastore.

Here is the example `Process` method implemented by the [ScyllaProcessor][],
the [ScyllaDB][] processor backend:

```golang
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
```

As you can see, it's a direct call to the ScyllaDB driver.

[scyllaprocessor]: ./pkg/processing/scylla.go

### Unit tests

As mentioned above, the unit test covers the data conversion/verification
code in [processing.Scan][scan-type].

Here is the output of the Go unit test:

```bash
$ go test ./pkg/processing/ -v
=== RUN   TestScanValidateIP
=== RUN   TestScanValidateIP/Valid_IPv4_address
=== RUN   TestScanValidateIP/Invalid_IPv4_address
=== RUN   TestScanValidateIP/Valid_IPv6_address
=== RUN   TestScanValidateIP/Invalid_IPv6_address
--- PASS: TestScanValidateIP (0.00s)
    --- PASS: TestScanValidateIP/Valid_IPv4_address (0.00s)
    --- PASS: TestScanValidateIP/Invalid_IPv4_address (0.00s)
    --- PASS: TestScanValidateIP/Valid_IPv6_address (0.00s)
    --- PASS: TestScanValidateIP/Invalid_IPv6_address (0.00s)
=== RUN   TestValidScanValidatePort
=== RUN   TestValidScanValidatePort/Valid_port
=== RUN   TestValidScanValidatePort/Invalid_port
--- PASS: TestValidScanValidatePort (0.00s)
    --- PASS: TestValidScanValidatePort/Valid_port (0.00s)
    --- PASS: TestValidScanValidatePort/Invalid_port (0.00s)
=== RUN   TestValidScanValidateData
=== RUN   TestValidScanValidateData/Valid_v1_data
=== RUN   TestValidScanValidateData/Invalid_v1_data_type
=== RUN   TestValidScanValidateData/Invalid_v1_data_encoding
=== RUN   TestValidScanValidateData/Valid_v2_data
=== RUN   TestValidScanValidateData/Invalid_v2_data_type
=== RUN   TestValidScanValidateData/Unsupported_data_version
--- PASS: TestValidScanValidateData (0.00s)
    --- PASS: TestValidScanValidateData/Valid_v1_data (0.00s)
    --- PASS: TestValidScanValidateData/Invalid_v1_data_type (0.00s)
    --- PASS: TestValidScanValidateData/Invalid_v1_data_encoding (0.00s)
    --- PASS: TestValidScanValidateData/Valid_v2_data (0.00s)
    --- PASS: TestValidScanValidateData/Invalid_v2_data_type (0.00s)
    --- PASS: TestValidScanValidateData/Unsupported_data_version (0.00s)
PASS
ok      github.com/censys/scan-takehome/pkg/processing  0.009s
```

It covers the majority of the scanned data field with the various input
data.

### End-to-end verification

Here is the end-to-end verification with the small set of the scanned data.

To make the verification easy, we modified the `scanner` instance to inject
small set of the scanned data.

Here is the diff of the [scanner][] used for this verification.

[scanner]: ./cmd/scanner/main.go

```bash
$ git diff cmd/scanner/
diff --git a/cmd/scanner/main.go b/cmd/scanner/main.go
index 3623ef7..f1bff21 100644
--- a/cmd/scanner/main.go
+++ b/cmd/scanner/main.go
@@ -32,8 +32,8 @@ func main() {
        for range time.Tick(time.Second) {

                scan := &scanning.Scan{
-                       Ip:        fmt.Sprintf("1.1.1.%d", rand.Intn(255)),
-                       Port:      uint32(rand.Intn(65535)),
+                       Ip:        fmt.Sprintf("1.1.1.%d", rand.Intn(2)),
+                       Port:      uint32(rand.Intn(3)),
                        Service:   services[rand.Intn(len(services))],
                        Timestamp: time.Now().Unix(),
                }
```

This change makes the `scanner` generate only 18 variations of scanned data,
as in:

```
18 variations = 2 IPs x 3 ports x 3 services
```

Re-build the scanner image and restart the docker compose:

```
docker compose build scanner
docker compose down
docker compose up
```

Run the following command and observe the ScyllaDB table for the
extended period of time, e.g. 5 minutes.

```bash
while true; do \
  echo 'select * from censys.services;' | cqlsh localhost; \
  sleep 1; \
done
```

Here is the example screenshot of the above command.

As you can see, we can observe only 18 entries of scanned data with
the variations generated by the modified `scanner`.

```bash
 ip      | port | service | data                 | timestamp
---------+------+---------+----------------------+---------------------------------
 1.1.1.0 |    0 |     DNS | service response: 32 | 2024-12-16 01:54:23.000000+0000
 1.1.1.0 |    0 |    HTTP |  service response: 9 | 2024-12-16 01:54:17.000000+0000
 1.1.1.0 |    0 |     SSH | service response: 19 | 2024-12-16 01:54:35.000000+0000
 1.1.1.0 |    1 |     DNS | service response: 11 | 2024-12-16 01:54:34.000000+0000
 1.1.1.0 |    1 |    HTTP | service response: 36 | 2024-12-16 01:54:04.000000+0000
 1.1.1.0 |    1 |     SSH | service response: 87 | 2024-12-16 01:54:33.000000+0000
 1.1.1.0 |    2 |     DNS |  service response: 0 | 2024-12-16 01:54:31.000000+0000
 1.1.1.0 |    2 |    HTTP | service response: 45 | 2024-12-16 01:54:30.000000+0000
 1.1.1.0 |    2 |     SSH | service response: 85 | 2024-12-16 01:54:24.000000+0000
 1.1.1.1 |    0 |     DNS | service response: 82 | 2024-12-16 01:54:36.000000+0000
 1.1.1.1 |    0 |    HTTP | service response: 46 | 2024-12-16 01:54:21.000000+0000
 1.1.1.1 |    0 |     SSH | service response: 57 | 2024-12-16 01:54:28.000000+0000
 1.1.1.1 |    1 |     DNS | service response: 39 | 2024-12-16 01:54:02.000000+0000
 1.1.1.1 |    1 |    HTTP | service response: 12 | 2024-12-16 01:53:52.000000+0000
 1.1.1.1 |    1 |     SSH | service response: 82 | 2024-12-16 01:54:16.000000+0000
 1.1.1.1 |    2 |     DNS | service response: 51 | 2024-12-16 01:54:32.000000+0000
 1.1.1.1 |    2 |    HTTP | service response: 53 | 2024-12-16 01:54:13.000000+0000
 1.1.1.1 |    2 |     SSH | service response: 75 | 2024-12-16 01:53:35.000000+0000

(18 rows)
```

### Final thoughts

First of all, I really enjoyed this take home quiz. Its well structured and
clean source code organization made me happy.  And the docker compose
based local development environment gave me a joy.

And also, I like the design of the GCP Pub/Sub Go package.  The API is
well-thought and clean.  It abstracts the details away from the application
developers.  Good job, Google!

As a next step, I want to explore the GCP Spanner based processor backend
and see if it's easily swappable with the processing.Processor interface.

Thank you, again, and your PR to my solution is more than welcome. :)

Happy hacking!
