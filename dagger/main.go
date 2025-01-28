package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"dagger.io/dagger/dag"
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
		WithMountedCache("/target", dag.CacheVolume("rust-target-001")).
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

	if err := Lint(ctx, baseImage); err != nil {
		handleError(err)
	}

	if err := Test(ctx, baseImage); err != nil {
		handleError(err)
	}

	if err := Coverage(ctx, baseImage); err != nil {
		handleError(err)
	}
}

// enforce format job
func Format(ctx context.Context, baseImage *dagger.Container) error {
	format := baseImage.
		WithExec([]string{"rustup", "component", "add", "rustfmt"}).
		WithExec([]string{"cargo", "fmt", "--check"})

	out, err := format.Stdout(ctx)
	if err != nil {
		return err
	}

	fmt.Println(out)

	return nil
}

// linting job
func Lint(ctx context.Context, baseImage *dagger.Container) error {
	clippy := baseImage.
		WithEnvVariable("SQLX_OFFLINE", "true").
		WithExec([]string{"rustup", "component", "add", "clippy"}).
		WithExec([]string{"cargo", "clippy", "--", "-D", "warnings"})

	out, err := clippy.Stdout(ctx)
	if err != nil {
		return err
	}

	fmt.Println(out)

	return nil
}

// test job
func Test(ctx context.Context, baseImage *dagger.Container) error {
	test := baseImage.
		WithExec([]string{"cargo", "sqlx", "prepare", "--workspace", "--check"}).
		WithExec([]string{"cargo", "test"})

	out, err := test.Stdout(ctx)
	if err != nil {
		return err
	}

	fmt.Println(out)

	return nil
}

// code coverage job
func Coverage(ctx context.Context, baseImage *dagger.Container) error {
	coverage := baseImage.
		WithExec([]string{"cargo", "install", "cargo-llvm-cov", "--locked"}).
		WithExec([]string{"cargo", "llvm-cov", "--all-features", "--workspace"})

	out, err := coverage.Stdout(ctx)
	if err != nil {
		return err
	}

	fmt.Println(out)

	return nil
}

func handleError(err error) {
	if err != nil {
		fmt.Println("error:", err)
		panic(err)
	}
}
