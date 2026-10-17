# Week 5 — CI pipeline with GitHub Actions

[![CI](https://github.com/chalshik/devops-week5-ci/actions/workflows/ci.yml/badge.svg)](https://github.com/chalshik/devops-week5-ci/actions/workflows/ci.yml)

The Go HTTP service from Week 3, now with unit tests and a GitHub Actions
pipeline that lints, tests (on a Go version matrix), and builds the Docker image
on every push, pull request, and manual dispatch.

GitHub only runs workflows from `.github/workflows/` at a repository root, so
this folder is also published as its own repository, where the pipeline runs:

- Repository: <https://github.com/chalshik/devops-week5-ci>
- Workflow runs: <https://github.com/chalshik/devops-week5-ci/actions/workflows/ci.yml>

See [CI.md](CI.md) for what each job does, the triggers, and how caching helps.

## Layout

| File | Purpose |
| --- | --- |
| `main.go` | HTTP server: `/` greeting, `/healthz` JSON health endpoint, `-healthcheck` probe flag |
| `main_test.go` | Five unit tests using `net/http/httptest` |
| `Dockerfile` | Multi-stage build: `golang:1.23-alpine` builder, `alpine:3.20` non-root runtime |
| `.github/workflows/ci.yml` | `lint` → `test` (matrix) → `build` pipeline |
| `CI.md` | Pipeline explanation |

## Pipeline

```
push / pull_request / workflow_dispatch
        │
        ▼
     lint ── gofmt, go vet
        │
        ▼
     test ── matrix: Go 1.23, Go 1.24 ── artifact: test-results-go<ver>
        │
        ▼
    build ── docker build (gha cache) ── smoke test /healthz ── artifact: image-digest
```

## Run locally

```sh
gofmt -l .
go vet ./...
go test -race -v ./...

docker build -t devops-week5:ci .
docker run -d --rm --name w5 -p 8080:8080 devops-week5:ci
curl -fsS localhost:8080/healthz
curl -fsS localhost:8080/
docker rm -f w5
```

Local results:

```
--- PASS: TestHealthz (0.00s)
--- PASS: TestRoot (0.00s)
--- PASS: TestUnknownPathReturns404 (0.00s)
--- PASS: TestPortFromEnv (0.00s)
--- PASS: TestHealthcheck (0.00s)
PASS
ok  	devops-week5

{"status":"ok"}
Hello from the DevOps week 5 container!
Up 12 seconds (healthy)
```
