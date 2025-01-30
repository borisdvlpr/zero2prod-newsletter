package main

import (
	"fmt"

	"dagger/dagger/internal/dagger"
)

// Constructor
func New(
	source *dagger.Directory,
) *Dagger {
	return &Dagger{
		Source:           source,
		SqlxVersion:      "0.7.4",
		SqlxFeatures:     "rustls,postgres",
		DatabaseHost:     "postgres",
		DatabaseUser:     "postgres",
		DatabasePassword: "password",
		DatabaseName:     "newsletter",
	}
}

type Dagger struct {
	Source           *dagger.Directory
	SqlxVersion      string
	SqlxFeatures     string
	DatabaseHost     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
}

// Build a ready-to-use development environment
func (m *Dagger) BuildEnv() *dagger.Container {
	// setup database container
	postgres := dag.Container().From("postgres:16-bookworm").
		WithEnvVariable("POSTGRES_PASSWORD", m.DatabasePassword).
		WithExposedPort(5432).
		AsService(dagger.ContainerAsServiceOpts{UseEntrypoint: true})

	// format version and features to use constants
	sqlxVersion := fmt.Sprintf("--version=%s", m.SqlxVersion)
	sqlxFeatures := fmt.Sprintf("--features=%s", m.SqlxFeatures)

	// create base container
	return dag.Container().From("rust:slim-bookworm").
		WithServiceBinding(m.DatabaseHost, postgres).
		WithDirectory("/hello-rust", m.Source).
		WithWorkdir("/hello-rust").
		WithMountedCache("/target", dag.CacheVolume("rust-target")).
		//WithEnvVariable("CARGO_TERM_COLOR", CARGO_TERM_COLOR).
		WithEnvVariable("SQLX_VERSION", m.SqlxVersion).
		WithEnvVariable("SQLX_FEATURES", m.SqlxFeatures).
		WithEnvVariable("POSTGRES_HOST", m.DatabaseHost).
		WithEnvVariable("POSTGRES_USER", m.DatabaseUser).
		WithEnvVariable("POSTGRES_PASSWORD", m.DatabasePassword).
		WithEnvVariable("POSTGRES_DB", m.DatabaseName).
		WithEnvVariable("DATABASE_URL", "postgres://postgres:password@postgres:5432/newsletter").
		WithEnvVariable("APP_ENVIRONMENT", "ci").
		WithEnvVariable("SKIP_DOCKER", "true").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "lld", "clang", "postgresql-client", "-y"}).
		WithExec([]string{"cargo", "install", "sqlx-cli", sqlxVersion, sqlxFeatures, "--no-default-features", "--locked"}).
		WithExec([]string{"./scripts/init_db.sh"})
}
