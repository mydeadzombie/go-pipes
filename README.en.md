## go-pipes

Lightweight pipeline graph library configured via YAML. The demo computes MD5 hashes of files in a directory with configurable parallelism.

### Quick start

```bash
# Run example
go run ./examples/md5
go run ./examples/md5 -dir=/path/to/dir -parallelism=20
go run ./examples/md5 -pipeline=examples/md5/pipeline.yml
```

### YAML schema

```yaml
nodes:
  - id: walker
    type: file_walker
    config:
      dir: "."
  - id: hasher
    type: md5_hasher
    config:
      workers: 10
  - id: printer
    type: printer
    config:
      quiet: false

edges:
  - from: walker.files
    to: hasher.paths
    buffer: 256
  - from: hasher.results
    to: printer.in
    buffer: 0
```

### Writing YAML pipelines

- **nodes**: list of processing units with unique `id`, `type`, and `config`.
  - `file_walker`:
    - emits file paths on port `files`
    - config: `dir` (string|list, can be multiple roots), `workers` (int, default 1)
  - `md5_hasher`:
    - input `paths` (string), output `results` (object with `Path`, `Sum`, `Err`)
    - config: `workers` (int, default 10)
  - `printer`:
    - input `in`
    - config: `quiet` (bool, default false), `workers` (int, default 1). Output lines are prefixed with `worker=<id>` when `workers>1`.
  - `file_sink`:
    - input `in`
    - config: `path` (string), `append` (bool), `workers` (int). When `workers>1` the output file is grouped by worker sections.
  - `tee`:
    - input `in`, outputs `out1`, `out2` — duplicates the stream into two directions.
  - `stdin_source`:
    - emits a single path to port `paths` read from stdin
    - config: `prompt` (string), `allowEmpty` (bool)
- **edges**: connections `from: <node>.<outPort>`, `to: <node>.<inPort>`, optional `buffer` (int, default 0).
- **rules**:
  - Node IDs must be unique.
  - Edge endpoints must be in `node.port` format.

Example:

```yaml
nodes:
  - id: walker
    type: file_walker
    config:
      dir: "/data"
  - id: hasher
    type: md5_hasher
    config:
      workers: 20
  - id: printer
    type: printer
    config:
      quiet: false
edges:
  - from: walker.files
    to: hasher.paths
    buffer: 256
  - from: hasher.results
    to: printer.in
```

<details>
<summary>Extended example (tee to console and file):</summary>

```yaml
nodes:
  - id: walker
    type: file_walker
    config:
      workers: 1
      dir:
        - "/bin"
        - "/app"
  - id: hasher
    type: md5_hasher
    config:
      workers: 10
  - id: printer
    type: printer
    config:
      workers: 3
      quiet: false
  - id: tee
    type: tee
    config: {}
  - id: fileout
    type: file_sink
    config:
      path: "/app/md5.txt"
      append: false

edges:
  - from: walker.files
    to: hasher.paths
    buffer: 256
  - from: hasher.results
    to: tee.in
    buffer: 0
  - from: tee.out1
    to: printer.in
    buffer: 0
  - from: tee.out2
    to: fileout.in
    buffer: 0
```
</details>

### Notes

- Default parallelism is 10.
- Backpressure is enforced by edge channel buffers.

### Docker

```bash
# Build example
docker build -t go-pipes-md5 -f examples/md5/Dockerfile .
docker run --rm -v $(pwd):/data go-pipes-md5 -pipeline=/pipeline.yml -dir=/data
```

### Docker Compose

```bash
# Build and Run example with docker compose
docker compose -f examples/md5/docker-compose.yml up --build
```

### Interactive example (stdin)

```bash
go run ./examples/md5 -pipeline=examples/md5/pipeline.stdin.yml
# or without waiting for user input:
echo "/path/to/file" | go run ./examples/md5 -pipeline=examples/md5/pipeline.stdin.yml
```


