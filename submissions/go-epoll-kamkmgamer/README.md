# Go benchmax: epoll + pgx pipelines

**Front: direct**

Language: Go 1.27.1  
Runtime: custom Linux epoll reactor  
PostgreSQL driver: pgx/v5 5.11.0  
Submission: kamkmgamer

This submission is tuned for the challenge's eight-core, direct-to-client benchmark. It keeps the
API and security behavior required by `SPEC.md`, while removing avoidable work from the hot path:

- four epoll loops with fixed request and input arenas, so idle keep-alive sockets do not own a
  goroutine or a growing buffer;
- separate read and write queues, prepared PostgreSQL statements, binary parameters/results,
  pipelined batches, and durable group commits;
- allocation-light JSON serialization, an allocation-free HS256 fast path with full signature and
  expiry checks, and a four-lane bcrypt implementation for cost-10 password work;
- concurrent `/api/tags` requests already waiting in one read batch share one fresh PostgreSQL
  execution, with an independent response buffer for every request;
- larger default request windows (`MAX_INFLIGHT=8192`, `MAX_QUEUED=32768`) to avoid queueing
  otherwise serviceable load at high concurrency.

The server still verifies every token on every request, hashes and checks passwords with bcrypt cost
10, commits writes before responding, and reads request data from PostgreSQL. No response or query
result is retained across requests.

## Build and run

```bash
sudo bash install.sh && bash build.sh
DATABASE_URL=postgres://conduit:conduit@127.0.0.1:5432/conduit \
  JWT_SECRET=multicore-server-challenge HOST=0.0.0.0 PORT=80 bash start.sh
```

The implementation accepts the tuning variables documented in `src/main.go`, `src/db.go`, and
`src/reactor_linux.go`; the defaults target the benchmark box and can be overridden for a local run.
