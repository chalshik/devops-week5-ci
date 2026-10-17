# CI pipeline

The workflow lives in `.github/workflows/ci.yml` and runs three jobs in order:
`lint` → `test` → `build`. Each job depends on the previous one through `needs`,
so a failure stops the pipeline early and nothing gets built from broken code.

## Triggers

| Event | When it fires |
| --- | --- |
| `push` | Every push to any branch |
| `pull_request` | Every PR opened, updated, or reopened, so changes are checked before merge |
| `workflow_dispatch` | Manual run from the Actions tab ("Run workflow") |

`concurrency` cancels an in-progress run on the same ref when a newer commit
arrives, so only the latest commit uses runner time. The workflow token is
limited to `contents: read`.

## Jobs

### `lint`

- Sets up Go 1.23 with the module and build cache enabled.
- `gofmt -l .` lists any file that is not formatted; if the list is not empty,
  the step exits 1.
- `go vet ./...` catches suspicious code, for example bad `Printf` arguments,
  unreachable code, or copied locks.

### `test`

- Runs as a **matrix** on Go `1.23` and `1.24`, so the code is checked on the
  version in `go.mod` and on the next release.
- `fail-fast: true`: if one matrix leg fails, the other is cancelled.
- `go test -race -v -coverprofile=...` runs the unit tests in `main_test.go`
  with the race detector on. `set -o pipefail` makes sure the exit code of
  `go test` survives the `tee`, so **a failing test fails the job** and the
  whole workflow.
- The test log and coverage profile are uploaded as the
  `test-results-go<version>` artifact. `if: always()` uploads them even when
  tests fail, which is when they are most useful.

Tests:

| Test | Checks |
| --- | --- |
| `TestHealthz` | `/healthz` returns 200, JSON content type, `{"status":"ok"}` |
| `TestRoot` | `/` returns 200 and the exact greeting |
| `TestUnknownPathReturns404` | Unknown paths return 404 |
| `TestPortFromEnv` | `PORT` env var is respected, default is 8080 |
| `TestHealthcheck` | The `-healthcheck` probe returns 0 for healthy and 1 for a non-200 |

### `build`

- Builds the multi-stage Docker image with Buildx and loads it into the runner.
- Smoke test: starts the container and polls `/healthz` for up to 30 seconds.
  If it never answers, the container logs are printed and the job fails.
- Writes the image ID (content digest) to `image-digest.txt` and uploads it as
  the `image-digest` artifact.

## Caching

| Cache | Where | What it saves |
| --- | --- | --- |
| Go modules and build cache | `actions/setup-go` with `cache: true`, keyed on `go.sum`/`go.mod` | Downloaded modules (`~/go/pkg/mod`) and compiled packages (`~/.cache/go-build`) are restored, so `go vet` and `go test` only recompile what changed |
| Docker layer cache | `cache-from/cache-to: type=gha` in `docker/build-push-action` | Base images and the `go mod download` layer are reused between runs; only the layers after `COPY main.go` rebuild on a code change |

On a warm cache the `lint` and `test` jobs skip module downloads and most
compilation, and the image build reuses every layer except the final compile.

## Pinned actions

Every action is pinned to a major version: `actions/checkout@v4`,
`actions/setup-go@v5`, `actions/upload-artifact@v4`,
`docker/setup-buildx-action@v3`, `docker/build-push-action@v6`.

## Secrets

The pipeline needs no credentials. The image is only built and tested on the
runner, never pushed, so no registry token is required. If one is added later,
it goes into repository secrets and is referenced as `${{ secrets.NAME }}`.

## Proving that tests fail the build

Change the greeting in `main.go`, for example `week 5` → `week 9`, and push.
`TestRoot` fails, `go test` exits 1, the `test` job goes red, and `build` is
skipped. Locally:

```sh
sed -i '' 's/week 5 container/week 9 container/' main.go
go test ./...
echo $?
```

The last command prints `1`.
