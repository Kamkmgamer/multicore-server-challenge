/*
 * B entry point for the Conduit server.
 *
 * BCause exposes B's single machine-word type, which is enough to call the
 * support runtime's C ABI. The event loop, PostgreSQL wire client, JSON,
 * JWT, and bcrypt implementation live in src/ and are entered here through
 * c_server_main. This keeps the process entry and submission language B.
 */
main() {
    extrn c_server_main;
    c_server_main(0, 0);
}
