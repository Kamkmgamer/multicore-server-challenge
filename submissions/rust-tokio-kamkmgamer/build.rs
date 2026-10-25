fn main() {
    println!("cargo:rerun-if-changed=native");
    if std::env::var_os("CARGO_FEATURE_NATIVE_BCRYPT").is_some()
        && std::env::var("CARGO_CFG_TARGET_OS").as_deref() == Ok("linux")
    {
        cc::Build::new()
            .file(
                if std::env::var_os("CARGO_FEATURE_BATCH_BCRYPT").is_some() {
                    "native/bcrypt-batch.c"
                } else {
                    "native/crypt-bcrypt.c"
                },
            )
            .include("native")
            .std("c11")
            .opt_level(3)
            // The benchmark compiles and runs on the same x86_64 host. Keep
            // native bcrypt on the same instruction-set path as the Rust code.
            .flag_if_supported("-march=native")
            .flag_if_supported("-mtune=native")
            .compile("conduit_bcrypt");
    }
}
