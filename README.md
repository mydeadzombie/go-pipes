## go-pipes

Легковесная библиотека для построения графов-пайплайнов, конфигурируемых через YAML.
Демо вычисляет MD5‑хеши файлов в директории с настраиваемым параллелизмом.

### Быстрый старт

```bash
# Run example
go run ./examples/md5
go run ./examples/md5 -dir=/path/to/dir -parallelism=20
go run ./examples/md5 -pipeline=examples/md5/pipeline.yml
```

### Схема YAML

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

### Как писать YAML‑пайплайны

- **nodes**: список узлов обработки с уникальными `id`, `type` и `config`.
  - `file_walker`:
    - выводит в порт `files` (строковые пути к файлам)
    - конфиг: `dir` (string|list, можно несколько директорий), `workers` (int, по умолчанию 1)
  - `md5_hasher`:
    - вход `paths` (string), выход `results` (объект с полями `Path`, `Sum`, `Err`)
    - конфиг: `workers` (int, необязательный, по умолчанию 10)
  - `printer`:
    - вход `in`
    - конфиг: `quiet` (bool, по умолчанию false), `workers` (int, по умолчанию 1). Вывод префиксируется `worker=<id>`.
  - `file_sink`:
    - вход `in`
    - конфиг: `path` (string), `append` (bool), `workers` (int). При `workers>1` вывод группируется секциями по worker.
  - `tee`:
    - вход `in`, выходы `out1`, `out2` — дублирует поток на два направления.
  - `stdin_source`:
    - выводит один путь в порт `paths`, читая строку из stdin
    - конфиг: `prompt` (string), `allowEmpty` (bool)
- **edges**: соединения вида `from: <node>.<outPort>`, `to: <node>.<inPort>`, опционально `buffer` (int, по умолчанию 0).
- **правила**:
  - Идентификаторы узлов должны быть уникальны.
  - Концы рёбер должны быть в формате `node.port`.

Пример:

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
<summary>Дополнительный пример (дублирование результатов в консоль и файл):</summary>

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

### Заметки

- Параллелизм по умолчанию: 10.
- Backpressure соблюдается за счёт буферов каналов на рёбрах.

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

### Интерактивный пример (stdin)

```bash
go run ./examples/md5 -pipeline=examples/md5/pipeline.stdin.yml
# или без ожидания ввода:
echo "/path/to/file" | go run ./examples/md5 -pipeline=examples/md5/pipeline.stdin.yml
```

