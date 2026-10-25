# B / BCause, kamkmgamer

Front: direct

This submission uses the B programming language for the process entrypoint.
`main.b` calls the C ABI support runtime because B has no maintained
PostgreSQL, io_uring, JWT, or bcrypt libraries. The support runtime is the
hand-written Conduit server in `src/`, while B owns the executable entry and
build path.

BCause is vendored under `bcause/` and compiles B to x86-64 assembly. The
support runtime provides direct HTTP/1.1 handling, PostgreSQL wire-protocol
access, prepared statements, JSON serialization, HS256 JWT checks, and
cost-10 bcrypt. It keeps the challenge's no-cache and durable-write rules.

`build.sh` compiles the B entrypoint, compiles the C support objects, and links
both with liburing, libxcrypt, pthreads, and the BCause standard library.
`start.sh` runs the resulting direct server. Runtime configuration comes from
`DATABASE_URL`, `JWT_SECRET`, `HOST`, and `PORT`; the support runtime also
accepts its normal tuning variables such as `THREADS`, `DB_CONNS`, and
`BCRYPT_THREADS`.

```sh
sudo bash install.sh
bash build.sh
DATABASE_URL=postgres://conduit:conduit@127.0.0.1:5432/conduit \
JWT_SECRET=multicore-server-challenge HOST=0.0.0.0 PORT=80 bash start.sh
```

The BCause compiler is MIT licensed. The Conduit support runtime and its
native bcrypt implementation retain their original MIT attribution.
