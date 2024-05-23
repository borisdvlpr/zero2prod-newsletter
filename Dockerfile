FROM rust:alpine

WORKDIR /app

RUN apk update && apk add lld clang

COPY . .

RUN cargo build --release

ENTRYPOINT ["./target/release/zero2prod"]
