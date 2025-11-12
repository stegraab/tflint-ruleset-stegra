# Repository Guidelines

## Project Structure & Module Organization
- `main.go`: Registers and serves the ruleset (`stegra`).
- `rules/`: Each rule in one file, e.g. `terraform_count_for_each_position.go` with a matching `*_test.go`.
- `.github/workflows/`: CI for build/test and tagged releases via GoReleaser.
- `Makefile`: Local build/test/install helpers.
- `go.mod`, `go.sum`: Go 1.25 module; plugin SDK dependencies.

## Build, Test, and Development Commands
- `make build`: Compile the plugin binary `tflint-ruleset-stegra`.
- `make test`: Run Go tests for all rules (`go test ./...`).
- `make install`: Build and place the binary in `~/.tflint.d/plugins/` for local use.
- Example local run:
  - Create `.tflint.hcl` with:
    ```hcl
    plugin "template" { enabled = true }
    ```
  - Run `tflint` in a Terraform project.

## Coding Style & Naming Conventions
- Language: Go. Use `go fmt ./...` and `go vet ./...` before committing.
- Rule files: snake_case, e.g. `terraform_count_for_each_position.go` and `*_test.go`.
- Types: `PascalCase`, e.g. `TerraformCountForEachPositionRule`.
- Rule metadata: `Name()` returns snake_case id; `Enabled()`/`Severity()` reflect defaults; set `Link()` if a doc exists.
- Register new rules in `main.go` inside `Rules: []tflint.Rule{ ... }`.

## Testing Guidelines
- Framework: Go `testing` with `tflint-plugin-sdk/helper`.
- Write table-driven tests covering:
  - Valid configurations (no issues).
  - Each violation with expected message, range, and file.
- Run locally with `make test`; optional coverage `go test -cover ./...`.

## Commit & Pull Request Guidelines
- Commits: concise imperative subject, e.g. "add rule: terraform_count_for_each_position".
- PRs must include:
  - Summary of rule/changes and rationale.
  - Examples of failing and passing HCL snippets.
  - Tests updated/added and CI green.
  - Linked issue (if applicable).

## Security & Release Tips
- Do not commit secrets. Releases are driven by tags `vX.Y.Z`.
- Requirements: TFLint >= 0.46, Go >= 1.25. Use `make install` for local plugin testing.
