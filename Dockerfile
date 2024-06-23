FROM rust:alpine

WORKDIR /app

RUN apk update && apk add lld clang -y

COPY . .

ENV SQLX_OFFLINE true
ENV APP_ENVIRONMENT production
RUN cargo build --release

ENTRYPOINT ["./target/release/zero2prod"]
