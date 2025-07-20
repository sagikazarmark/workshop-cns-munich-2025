package main

import (
	"context"

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
