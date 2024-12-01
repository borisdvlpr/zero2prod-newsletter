use sqlx::postgres::PgPoolOptions;
use std::io;
use std::net::TcpListener;
use zero2prod_newsletter::configuration::get_configuration;
use zero2prod_newsletter::email_client::EmailClient;
use zero2prod_newsletter::startup::build;
use zero2prod_newsletter::telemetry::{get_subscriber, init_subscriber};

#[tokio::main]
async fn main() -> Result<(), std::io::Error> {
    let subscriber = get_subscriber("zero2prod".into(), "info".into(), io::stdout);
    init_subscriber(subscriber);

    let configuration = get_configuration().expect("Failed to read configuration.");
    let server = build(configuration).await?;
    server.await?;

    Ok(())
}
