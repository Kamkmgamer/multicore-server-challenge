#!/usr/bin/env bash
# Builds the B entrypoint and links it to the hand-written support runtime.
set -euo pipefail
cd "$(dirname "$0")"
CC="${CC:-gcc}"
mkdir -p bin .build
make -C bcause CC="$CC" >/dev/null
# BCause emits a.s for -S. Compile that B-generated assembly with the same host toolchain.
./bcause/bcause -S main.b >/dev/null
"$CC" -c a.s -o .build/bmain.o
objs=()
for src in src/*.c; do
  name="$(basename "$src" .c)"
  obj=".build/${name}.o"
  "$CC" -std=gnu11 -O2 -march=native -pthread -Wall -Wextra -Wno-unused-parameter \
    -Dmain=c_server_main -c "$src" -o "$obj"
  objs+=("$obj")
done
"$CC" -nostartfiles -o bin/conduit .build/bmain.o "${objs[@]}" bcause/libb.a \
  -luring -lcrypt -lm -lpthread
printf 'built B entrypoint with %s\n' "$($CC --version | head -1)"
