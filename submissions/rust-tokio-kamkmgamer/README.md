# Rust / Tokio benchmax — kamkmgamer

Front: direct

Language: Rust
Framework/runtime: Tokio with a hand-written HTTP/1.1 server

This submission targets the multicore server challenge with Rust 1.99.0, Tokio 1.53, a hand-written HTTP/1.1
front end, a pipelined `tokio-postgres` client, mimalloc, native batched bcrypt, and direct JSON serialization.
The API implementation is based on the challenge's MIT-licensed optimized Rust implementation; the complete
source and its attribution are kept in this folder.

The hot path is tuned for the benchmark box:

- `build.sh` enables `-C target-cpu=native`; release builds use fat LTO, one codegen unit, panic abort, and a
  stripped binary.
- Direct mode avoids the Nginx hop and keeps HTTP/1.1 connections alive. Request and response buffers use
  thread-local free lists, and the parser accepts pipelined requests without per-request framework objects.
- Reads use prepared binary PostgreSQL statements, connection-load routing, one-snapshot read groups, and
  compatible page coalescing. Mutations use dedicated connections and durable grouped commits.
- Password work remains bcrypt cost 10, but independent jobs are batched through the native Openwall-compatible
  implementation so login/hash bursts do not stall the Tokio workers.
- HS256 signatures, expiry, and the requested user row are checked on every request. Canonical loadgen tokens use a
  strict allocation-free claims fast path; non-canonical signed tokens still go through full JSON validation. Query
  integers use a single checked scan. No response or data cache is retained between requests.

Build and run:

```sh
bash install.sh
bash build.sh
DATABASE_URL=postgres://conduit:conduit@127.0.0.1:5432/conduit \
JWT_SECRET=multicore-server-challenge HOST=0.0.0.0 PORT=80 bash start.sh
```

Runtime knobs from the source implementation remain available (`APP_THREADS`, `DB_CONNS`, `DB_WRITE_CONNS`,
`BCRYPT_THREADS`, `BCRYPT_BATCH_MS`, `WRITE_BATCH_US`, `READ_BATCH_SIZE`, `READ_BATCH_US`, and `READ_SQL_BATCH`).
The default profile is two Tokio workers, six read connections, two write connections, and one bcrypt worker on
the challenge's eight-core box.
