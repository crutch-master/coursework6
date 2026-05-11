FROM rust:1.93-alpine AS builder
RUN apk add --no-cache musl-dev
WORKDIR /app

COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo 'fn main() {}' > src/main.rs
RUN cargo build --release && rm -rf src

COPY src src
RUN touch src/main.rs && cargo build --release

FROM scratch
COPY --from=builder /app/target/release/worker /bin/worker
ENTRYPOINT ["/bin/worker"]
