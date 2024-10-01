use sqlx::postgres::PgPoolOptions;
use std::io;
use std::net::TcpListener;
use zero2prod_newsletter::configuration::get_configuration;
use zero2prod_newsletter::email_client::EmailClient;
use zero2prod_newsletter::startup::run;
use zero2prod_newsletter::telemetry::{get_subscriber, init_subscriber};

#[tokio::main]
async fn main() -> Result<(), std::io::Error> {
    let subscriber = get_subscriber("zero2prod".into(), "info".into(), io::stdout);
    init_subscriber(subscriber);

    let configuration = get_configuration().expect("Failed to read configuration.");

    let connection_pool = PgPoolOptions::new().connect_lazy_with(configuration.database.with_db());
    let address = format!(
        "{}:{}",
        configuration.application.host, configuration.application.port
    );

    // build an 'EmailClient' using 'configuration'
    let sender_email = configuration
        .email_client
        .sender()
        .expect("Invalid sender email address");
    let email_client = EmailClient::new(
        configuration.email_client.base_url,
        sender_email,
        configuration.email_client.authorization_token,
    );

    let listener = TcpListener::bind(address)?;
    run(listener, connection_pool, email_client)?.await
}
