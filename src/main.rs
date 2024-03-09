use zero2prod_newsletter::run;

#[tokio::main]
async fn main() -> Result<(), std::io::Error> {
    run().await
}
