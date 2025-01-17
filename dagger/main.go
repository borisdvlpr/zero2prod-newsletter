package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

const (
	CARGO_TERM_COLOR  = "always"
	SQLX_VERSION      = "0.7.4"
	SQLX_FEATURES     = "rustls,postgres"
	POSTGRES_HOST     = "postgres"
	POSTGRES_USER     = "postgres"
	POSTGRES_PASSWORD = "password"
	POSTGRES_DB       = "newsletter"
)

func main() {
	// create a shared context
	ctx := context.Background()

	// create a dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		handleError(err)
	}

	defer func(client *dagger.Client) {
		err := client.Close()
		if err != nil {
			handleError(err)
		}
	}(client)

	// format version and features to use constants
	sqlxVersion := fmt.Sprintf("--version=%s", SQLX_VERSION)
	sqlxFeatures := fmt.Sprintf("--features=%s", SQLX_FEATURES)

	// setup database container
	postgres := client.Container().From("postgres:16-bookworm").
		WithEnvVariable("POSTGRES_PASSWORD", POSTGRES_PASSWORD).
		WithExposedPort(5432).
		AsService(dagger.ContainerAsServiceOpts{UseEntrypoint: true})

	// setup base container image with rust and necessary dependencies, used on all stages
	baseImage := client.Container().From("rust:slim-bookworm").
		WithServiceBinding(POSTGRES_HOST, postgres).
		WithDirectory("/hello-rust", client.Host().Directory(".")).
		WithWorkdir("/hello-rust").
		WithEnvVariable("CARGO_TERM_COLOR", CARGO_TERM_COLOR).
		WithEnvVariable("SQLX_VERSION", SQLX_VERSION).
		WithEnvVariable("SQLX_FEATURES", SQLX_FEATURES).
		WithEnvVariable("POSTGRES_HOST", POSTGRES_HOST).
		WithEnvVariable("POSTGRES_USER", POSTGRES_USER).
		WithEnvVariable("POSTGRES_PASSWORD", POSTGRES_PASSWORD).
		WithEnvVariable("POSTGRES_DB", POSTGRES_DB).
		WithEnvVariable("DATABASE_URL", "postgres://postgres:password@postgres:5432/newsletter").
		WithEnvVariable("APP_ENVIRONMENT", "ci").
		WithEnvVariable("SKIP_DOCKER", "true").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "lld", "clang", "postgresql-client", "-y"}).
		WithExec([]string{"cargo", "install", "sqlx-cli", sqlxVersion, sqlxFeatures, "--no-default-features", "--locked"}).
		WithExec([]string{"./scripts/init_db.sh"})

	// run the stages of the pipeline
	if err := Format(ctx, baseImage); err != nil {
		handleError(err)
	}
}

// enforce format job
func Format(ctx context.Context, baseImage *dagger.Container) error {
	format := baseImage.
		WithExec([]string{"rustup", "component", "add", "rustfmt"}).
		WithExec([]string{"cargo", "fmt", "--check"})

	_, err := format.ExitCode(ctx)
	if err != nil {
		return err
	}

	return nil
}

func handleError(err error) {
	if err != nil {
		fmt.Println("error:", err)
		panic(err)
	}
}
