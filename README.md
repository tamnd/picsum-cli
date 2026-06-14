# picsum

A command line for [Lorem Picsum](https://picsum.photos/) — random high-quality
placeholder images with metadata.

`picsum` is a single pure-Go binary. It reads public Lorem Picsum data
over plain HTTPS, shapes it into clean records, and prints output that pipes
into the rest of your tools. No API key, nothing to run alongside it.

The same package is also a [resource-URI driver](#use-it-as-a-resource-uri-driver),
so a host program like [ant](https://github.com/tamnd/ant) can address
picsum as `picsum://` URIs.

## Install

```bash
go install github.com/tamnd/picsum-cli/cmd/picsum@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/picsum-cli/releases), or run
the container image:

```bash
docker run --rm ghcr.io/tamnd/picsum:latest --help
```

## Usage

```bash
picsum list                        # list all images (20 per page)
picsum list --page 2 --limit 5    # page 2, 5 per page
picsum image 42                    # get info about image #42
picsum image 0 -o json             # as JSON, ready for jq
picsum --help                      # the whole command tree
```

Every command shares one output contract: `-o table|json|jsonl|csv|tsv|url|raw`,
`--fields` to pick columns, `--template` for a custom line, and `-n` to limit.
The default adapts to where output goes (a table on a terminal, JSONL in a
pipe), so the same command reads well by hand and parses cleanly downstream.

## Serve it

The same operations are available over HTTP and as an MCP tool set for agents,
with no extra code:

```bash
picsum serve --addr :7777    # GET /v1/list and /v1/image/<id>  return NDJSON
picsum mcp                   # speak MCP over stdio
```

## Use it as a resource-URI driver

`picsum` registers a `picsum` domain the way a program registers a
database driver with `database/sql`. A host enables it with one blank import:

```go
import _ "github.com/tamnd/picsum-cli/picsum"
```

Then [ant](https://github.com/tamnd/ant) (or any program that links the package)
dereferences `picsum://` URIs without knowing anything about picsum:

```bash
ant get picsum://image/42    # fetch image #42 metadata
ant url picsum://image/42    # the live https URL
```

## Development

```
cmd/picsum/   thin main: hands cli.NewApp to kit.Run
cli/          assembles the kit App from the picsum domain
picsum/       the library: HTTP client, data models, and domain.go (the driver)
docs/         tago documentation site
```

```bash
make build      # ./bin/picsum
make test       # go test ./...
make vet        # go vet ./...
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a
cosign signature:

```bash
git tag v0.1.0
git push --tags
```

The Homebrew and Scoop steps self-disable until their tokens exist, so the first
release works with no extra secrets.

## License

Apache-2.0. See [LICENSE](LICENSE).
