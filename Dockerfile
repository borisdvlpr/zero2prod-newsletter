# build stage
FROM rust:slim AS builder

WORKDIR /app

RUN apt update && apt install lld clang -y

COPY . .
ENV SQLX_OFFLINE true
RUN cargo build --release

# runtime stage
FROM debian:bookworm-slim AS runtime

WORKDIR /app

# install openssl - dynamically linked to some dependenices
# install ca-certificates - needed to verify tls certificates for HTTPS
RUN apt-get update -y \
    && apt-get install -y --no-install-recommends openssl ca-certificates \
    # clean up
    && apt-get autoremove -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/zero2prod zero2prod
COPY config config
ENV APP_ENVIRONMENT production

ENTRYPOINT ["./zero2prod"]
