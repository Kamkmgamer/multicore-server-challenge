# Temporary optimization follow-up

**Cleanup candidate:** remove this file after the next optimization pass.

The optimized branch was validated on the Synteq 8-vCPU/16-GB target with the full 2.315 GB seed:

- Conformance: 197/197 passed.
- 10,000 users from the separate load host: 2,545 req/s, p95 156 ms, p99 268 ms, 0 errors.
- 15,000 users: 3,805 req/s, p95 170 ms, p99 288 ms, 0 errors.
- 18,000 users: 4,604 req/s, p95 167 ms, p99 294 ms, 0 errors.
- 20,000 users passed at 5,119 req/s, but the load host exceeded the benchmark's 75% CPU limit.
- 30,000 and 60,000-user attempts were load-generator failures caused by local ephemeral-port exhaustion.

The next pass should reduce the remaining per-article correlated work, especially favorites counts, tag arrays,
and comments, then repeat the official five-minute search with a clean separate load box.
