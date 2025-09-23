// A generated module for GoFile functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"dagger/go-file/internal/dagger"
)

type GoFile struct{}

func (m *GoFile) Build(source *dagger.Directory) *dagger.Container {
	return dag.Container().
		From("golang:1.25").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"go", "build", "./..."}).
		WithExec([]string{"go", "test", "./..."})
}

func (m *GoFile) Lint(source *dagger.Directory) /*(string, error)*/ *dagger.Container {
	return dag.Container().
		From("golangci/golangci-lint:latest").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"golangci-lint", "custom"}).
		WithExec([]string{"./build/custom-gcl"})
}
