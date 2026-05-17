# Terminal

Go API service with a local development workflow based on Docker Compose, Postgres, Redis, and Air live reload.

## Requirements

- Docker
- Docker Compose
- Go 1.26+ if you want to run or build on the host
- Air if you want host-only live reload

Install Air for host-only development:

```bash
go install github.com/air-verse/air@latest
```

You do not need Air installed on the host when using `docker compose up app`; the dev image includes it.

## Configuration

Local configuration is read from `.env`.

Create a local `.env` file with your own values. Keep `.env` out of version control.

```env
PORT=8080

POSTGRES_USER=<local-postgres-user>
POSTGRES_PASSWORD=<local-postgres-pass>
POSTGRES_DB=terminal
POSTGRES_PORT=5432

REDIS_PORT=6379

AUTH_SERVER_URL=http://localhost:4000/
AUTH_AUDIENCE=http://localhost:8080

DATABASE_URL=postgres://<local-postgres-user>:<local-postgres-pass>@localhost:5432/terminal?sslmode=disable
REDIS_ADDR=localhost:6379
```

`PORT` controls both the port the Go app binds to and the host port Docker maps to. A single value covers both — if you change it, it changes everywhere.

`AUTH_SERVER_URL` is the external auth service base URL. The API fetches signing keys from `${AUTH_SERVER_URL}/api/auth/jwks`. `AUTH_AUDIENCE` must match the audience configured by the auth service for this API.

When the app runs inside Compose, `docker-compose.yml` overrides connection hosts so containers talk over the Docker network:

```env
DATABASE_URL=postgres://<db-user>:<db-pass>@postgres:5432/<db-name>?sslmode=disable
REDIS_ADDR=redis:6379
```

That means:

- Use `localhost` when running the Go app on your host.
- Use `postgres` and `redis` when running the app inside Compose.

## Development With Docker Compose

Docker Compose is the primary development environment. The app service runs Air inside the container with the source tree mounted as a volume — file changes on the host trigger a rebuild inside the container automatically.

Start the full development stack:

```bash
docker compose up app
```

This starts:

- `terminal-api`: Go app running through Air
- `terminal-postgres`: Postgres 18
- `terminal-redis`: Redis 8 Alpine

Run detached:

```bash
docker compose up -d app
docker logs -f terminal-api
```

Rebuild the dev image after Dockerfile or dependency changes:

```bash
docker compose build app
docker compose up app
```

Stop and remove the Compose containers/network:

```bash
docker compose down
```

Remove containers and service images:

```bash
docker compose down --rmi all
```

Remove containers, service images, and named volumes:

```bash
docker compose down --rmi all --volumes
```

Only use `--volumes` when you are comfortable deleting local Postgres and Redis data.

## Live Reload

Air is configured in `.air.toml`.

It builds:

```bash
go build -buildvcs=false -o ./tmp/terminal-api ./cmd/api
```

It runs:

```bash
./tmp/terminal-api
```

Watched files include:

- `.go`
- `.tpl`
- `.tmpl`
- `.html`
- `.env`

Ignored directories include:

- `tmp`
- `vendor`
- `testdata`
- `.git`
- `.agents`
- `.codex`

Host-only reload, assuming Postgres and Redis are already running in Compose:

```bash
air -c .air.toml
```

Do not run Air on the host while `terminal-api` is also running — both will compete for the same port.

## Docker Images

The Dockerfile has separate targets:

- `air`: builds the Air binary
- `dev`: Alpine Go image with Air for live reload
- `builder`: compiles the app
- final `scratch`: production image with only the compiled binary

Compose builds the dev target and tags it as:

```text
terminal-app:dev
```

Do not expect the dev image to be tiny. It includes a Go toolchain so Air can rebuild the app inside the container.

Build the production scratch image:

```bash
docker build -t terminal-app .
```

Run the production image:

```bash
docker run --rm -p 8080:8080 terminal-app
```

## Important Image Naming

`terminal-app:dev` is the live-reload image.

```bash
docker compose up app
```

`terminal-app:latest` is only created when you explicitly build the production image:

```bash
docker build -t terminal-app .
```

Avoid running the dev image directly with plain `docker run`, because it expects the source tree and `.air.toml` to be mounted at `/app`.

If you really want to run the dev image manually:

```bash
docker run --rm -it \
  -v "$PWD:/app" \
  -w /app \
  terminal-app:dev
```

Usually, prefer Compose instead.

## Data Volumes

Compose uses named volumes:

- `terminal_pgdata`: Postgres data
- `terminal_redis_data`: Redis data

Postgres 18 is mounted at:

```yaml
pgdata:/var/lib/postgresql
```

This is intentional. The Postgres 18 Docker image expects the parent directory mount, not the older `/var/lib/postgresql/data` mount.

List volumes:

```bash
docker volume ls
```

Delete this project's data volumes:

```bash
docker volume rm terminal_pgdata terminal_redis_data
```

## Useful Commands

Check service status:

```bash
docker compose ps
```

Follow app logs:

```bash
docker logs -f terminal-api
```

Follow Postgres logs:

```bash
docker logs -f terminal-postgres
```

Connect to Postgres:

```bash
docker exec -it terminal-postgres psql -U <db-user> -d <db-name>
```

Run a quick database check:

```bash
docker exec terminal-postgres psql -U <db-user> -d <db-name> -c 'select current_user, current_database();'
```

Check image sizes:

```bash
docker images
```

Check Docker disk usage:

```bash
docker system df
```

Prune unused build cache:

```bash
docker builder prune
```

## Troubleshooting

### `open .air.toml: no such file or directory`

You are probably running the dev image directly without mounting the repo.

Use Compose:

```bash
docker compose up app
```

Or build and run the production scratch image:

```bash
docker build -t terminal-app .
docker run --rm -p 8080:8080 terminal-app
```

### Address already in use

The app and a local Air process are competing for the same port. Only one should run at a time.

If Compose is running, stop the local process. If developing on the host, stop the Compose app service first:

```bash
docker compose stop app
```

### Postgres keeps restarting

Check logs:

```bash
docker logs terminal-postgres
```

If the logs mention Postgres 18 and `/var/lib/postgresql/data`, confirm the compose volume target is:

```yaml
- pgdata:/var/lib/postgresql
```

### App cannot connect to Postgres

Use `localhost` from the host:

```env
DATABASE_URL=postgres://<db-user>:<db-pass>@localhost:5432/<db-name>?sslmode=disable
```

Use `postgres` inside Compose:

```env
DATABASE_URL=postgres://<db-user>:<db-pass>@postgres:5432/<db-name>?sslmode=disable
```

The Compose app service already overrides this for container use.

### Host build gets permission denied in `tmp`

This can happen if a container previously created `tmp` as root. Fix ownership:

```bash
sudo chown -R "$USER:$USER" tmp .cache
```

The current Compose app service runs as `${UID:-1000}:${GID:-1000}` to avoid recreating that issue.

### Dev image is larger than the production image

Expected.

The dev image contains Go and Air so it can rebuild on file changes. The production image is the final scratch stage and contains only the compiled binary.

Build production:

```bash
docker build -t terminal-app .
docker images terminal-app
```

![Alt](https://repobeats.axiom.co/api/embed/8ae105d80894602025afb33b0cac4aa38a956b4b.svg "Repobeats analytics image")
