# Rust / Tokio, kamkmgamer

Front: direct

This submission implements the Conduit API with Rust 1.99.0, Tokio 1.53.2,
the vendored tokio-postgres 0.7.18 client, mimalloc 0.1.52, and PostgreSQL 18.
All dependency versions are pinned in `Cargo.lock`.

The server uses a hand-written HTTP/1.1 parser and JSON writer. It keeps
connections alive, reuses request buffers, and sends responses without an
intermediate framework object. The direct front end avoids the Nginx hop.

Postgres work uses prepared binary statements, separate read and write pools,
and load-aware routing. Concurrent reads can share one execution and enrich
each distinct article once. Writes use grouped transactions and release a
response only after `COMMIT` completes. Failed groups roll back and retry
requests individually.

The bcrypt worker uses the Openwall-compatible cost-10 implementation in
`native/`. It batches up to four independent jobs, uses fresh salts, and keeps
the Tokio workers available for HTTP and database work. Every request still
checks its JWT signature and expiry, and every request reads its data from
Postgres. The server keeps no response, query-result, or token-verification
cache and makes no schema or server-wide database changes.

`build.sh` enables fat LTO, one codegen unit, panic abort, stripped symbols,
native Rust instructions, and native C instructions for bcrypt. Set
`RUSTFLAGS` to replace the native Rust flag when building on another host.

On the fixed eight-core benchmark box, the B runtime profile is two Tokio
workers, six read connections, two grouped write connections, one bcrypt
worker, and 32-request read coalescing. The Rust runtime derives those pool
sizes from the available CPUs, so local hosts can use their capacity.
The following settings remain overridable: `APP_THREADS`, `DB_CONNS`,
`DB_WRITE_CONNS`, `BCRYPT_THREADS`, `BCRYPT_BATCH_MS`, `WRITE_BATCH_US`,
`READ_BATCH_SIZE`, `READ_BATCH_US`, `READ_SQL_BATCH`, and
`READ_ROUTE_SLACK`, `READ_TAG_BATCH`, `READ_TAG_ROUTE`, and
`READ_COMMENT_SET`.

```sh
sudo bash install.sh
bash build.sh
DATABASE_URL=postgres://conduit:conduit@127.0.0.1:5432/conduit \
JWT_SECRET=multicore-server-challenge HOST=0.0.0.0 PORT=80 bash start.sh
```

The API and initial SQL are based on [cknutson12's MIT-licensed Rust
submission](https://github.com/arjaythedev/multicore-server-challenge/pull/1).
The native bcrypt source and its license are documented in `native/README.md`.

Validation in this workspace: the Rust release unit suite passed all 13 active
tests, and the Conduit conformance suite passed all 197 checks on a fresh
PostgreSQL 18 development seed. A 1,000-user local load smoke test completed
with zero request errors; the load generator used more than 75% CPU, so that
run is not a server score.
