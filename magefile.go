//go:build mage

// Mage build file for omnirouter.
// Powered by nava (https://github.com/vinaycharlie01/shroute/nava).
//
// The backend lives under backend/ (Hexagonal Architecture: domain,
// application, adapters, infrastructure); go.yaml configures every Go
// target below to run scoped to that directory.
//
// Usage:
//
//	go install github.com/magefile/mage@latest
//	mage -l          # list targets
//	mage build       # compile for current platform
//	mage test        # run unit tests
//	mage integration # run integration tests (testcontainers, requires Docker)
//	mage lint        # run golangci-lint
package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	dockermagex "github.com/vinaycharlie01/shroute/nava/mage/docker"
	gomagex "github.com/vinaycharlie01/shroute/nava/mage/golang"
	goreleasermagex "github.com/vinaycharlie01/shroute/nava/mage/goreleaser"
	nodejsmagex "github.com/vinaycharlie01/shroute/nava/mage/nodejs"
)

// init loads all YAML configs once before any target runs.
func init() {
	_ = gomagex.LoadConfig("go.yaml")
	_ = dockermagex.LoadConfig("docker.yaml")
	_ = goreleasermagex.LoadConfig("goreleaser.yaml")
	_ = nodejsmagex.LoadConfig("node.yaml")
}

// ---- Go targets --------------------------------------------------------

// Build compiles the omnirouter backend for the current platform (config: go.yaml → build).
func Build() error { return gomagex.Build() }

// Test runs the unit test suite (config: go.yaml → test).
func Test() error { return gomagex.Test() }

// Lint runs golangci-lint (config: go.yaml → lint).
func Lint() error { return gomagex.Lint() }

// Vet runs go vet (config: go.yaml → vet).
func Vet() error { return gomagex.Vet() }

// Setup downloads Go modules (config: go.yaml → setup).
func Setup() error { return gomagex.Setup() }

// Race runs tests with race detection (config: go.yaml → race).
func Race() error { return gomagex.Race() }

// Coverage runs tests with coverage profiling (config: go.yaml → coverage).
func Coverage() error { return gomagex.Coverage() }

// Bench runs benchmarks (config: go.yaml → bench).
func Bench() error { return gomagex.Bench() }

// Govulncheck runs govulncheck for vulnerability scanning (config: go.yaml → govulncheck).
func Govulncheck() error { return gomagex.Govulncheck() }

// Integration runs the integration test suite against live Postgres/Redis
// testcontainers. Requires Docker to be available on the host
// (config: go.yaml → integration; tests are guarded by the "integration"
// build tag).
func Integration() error { return gomagex.Integration() }

// BuildLinux cross-compiles for linux/amd64, linux/arm64, darwin/amd64, and
// darwin/arm64 (config: go.yaml → crossBuild).
func BuildLinux() error { return gomagex.CrossBuild() }

// Clean removes build artefacts.
func Clean() error {
	fmt.Println("cleaning backend/dist/")
	return os.RemoveAll("backend/dist")
}

// ---- Docker targets ----------------------------------------------------

// Docker namespace for container operations.
type Docker mg.Namespace

// Build builds a multi-platform container image (config: docker.yaml → buildxBuild).
func (Docker) Build() error { return dockermagex.BuildxBuild() }

// Push pushes the image to the registry (config: docker.yaml → push).
func (Docker) Push() error { return dockermagex.Push() }

// Login logs in to the container registry (config: docker.yaml → login).
func (Docker) Login() error { return dockermagex.Login() }

// ---- Frontend targets ---------------------------------------------------

// Frontend namespace for npm-script-driven frontend operations.
type Frontend mg.Namespace

// Setup installs frontend npm dependencies (config: node.yaml → setup).
func (Frontend) Setup() error { return nodejsmagex.Setup() }

// Dev runs the frontend dev server (config: node.yaml → dev).
func (Frontend) Dev() error { return nodejsmagex.Dev() }

// Build builds the frontend for production (config: node.yaml → build).
func (Frontend) Build() error { return nodejsmagex.Build() }

// Lint runs frontend linting (config: node.yaml → lint).
func (Frontend) Lint() error { return nodejsmagex.Lint() }

// Test runs frontend tests (config: node.yaml → test).
func (Frontend) Test() error { return nodejsmagex.Test() }

// ---- Release target ----------------------------------------------------

// Release creates a GitHub release via goreleaser (config: goreleaser.yaml).
func Release() error { return goreleasermagex.Release() }

// Snapshot creates a local snapshot build without publishing (config: goreleaser.yaml).
func Snapshot() error { return goreleasermagex.Snapshot() }
