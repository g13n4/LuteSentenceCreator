FROM rust:1-alpine3.23 AS builder

RUN APK ADD GIT

WORKDIR /app
RUN git clone https://github.com/g13n4/sudachi.rs.git .

RUN cargo build --release

FROM debian:bookworm-slim

RUN mkdir -p /shared

COPY --from=builder /app/target/release/sudachi /shared/
