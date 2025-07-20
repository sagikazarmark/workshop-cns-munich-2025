package main

import (
	"context"
	"fmt"

	"dagger/workshop/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

type Workshop struct {
	// +private
	Source *dagger.Directory
}

func New(
	// Project source directory.
	//
	// +defaultPath=/
	// +ignore=[".devenv", ".direnv", ".github", "build"]
	source *dagger.Directory,
) *Workshop {
	return &Workshop{
		Source: source,
	}
}

// Build the application.
func (m *Workshop) Build() *dagger.File {
	return dag.Go().
		WithPlatform("linux/amd64").
		Build(m.Source, dagger.GoBuildOpts{
			Trimpath: true,
		})
}

// Run tests.
func (m *Workshop) Test() *dagger.Container {
	return dag.Go().
		WithSource(m.Source).
		Exec([]string{"go", "test", "-v", "./..."})
}

// Run linter.
func (m *Workshop) Lint() *dagger.Container {
	return dag.GolangciLint().
		Run(m.Source, dagger.GolangciLintRunOpts{
			Verbose: true,
		})
}

// Run all checks.
func (m *Workshop) Check(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(func(ctx context.Context) error {
		_, err := m.Build().Sync(ctx)

		return err
	})

	p.Go(func(ctx context.Context) error {
		_, err := m.Test().Sync(ctx)

		return err
	})

	p.Go(func(ctx context.Context) error {
		_, err := m.Lint().Sync(ctx)

		return err
	})

	return p.Wait()
}

// Build a container image.
func (m *Workshop) BuildContainerImage() *dagger.Container {
	binary := m.Build()

	return dag.Container().
		From("alpine:3.21").
		WithExec([]string{"sh", "-c", "apk add --update --no-cache ca-certificates tzdata"}).
		WithFile("/usr/local/bin/app", binary).
		WithEntrypoint([]string{"app"})
}

// Run the application.
func (m *Workshop) Run() *dagger.Service {
	return m.BuildContainerImage().
		WithExposedPort(8080).
		AsService()
}

func (m *Workshop) ReleaseDummy(
	ctx context.Context,
	version string,
	username string,
) error {
	repository := fmt.Sprintf("ttl.sh/%s/workshop-kcd-czsk-2025:%s", username, version)

	_, err := m.BuildContainerImage().
		Publish(ctx, repository)
	if err != nil {
		return err
	}

	return nil
}

func (m *Workshop) Release(
	ctx context.Context,
	version string,
	registry string,
	repository string,
	username string,
	password *dagger.Secret,
) error {
	repository = fmt.Sprintf("%s/%s:%s", registry, repository, version)

	_, err := m.BuildContainerImage().
		WithRegistryAuth(registry, username, password).
		Publish(ctx, repository)
	if err != nil {
		return err
	}

	return nil
}
