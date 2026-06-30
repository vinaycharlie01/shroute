# .clinerules/workflows/

Project-specific Cline slash-command workflows. Each Markdown file here
(e.g. `deploy.md`) becomes a `/<filename>` command available in Cline for
this project only (as opposed to `~/Documents/Cline/Workflows/`, which is
global across projects).

Empty for now; add one here, for example, when the lint-first verification
sequence from `.clinerules/01-architecture-and-testing.md`
(`go build && go vet && gofmt -l . && golangci-lint run && go test ./...`)
is worth turning into a one-shot `/verify` workflow.
