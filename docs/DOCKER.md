# SSTorytime Docker Setup

Complete guide for running SSTorytime in Docker.

---

## Quick Start (see [The Docker quickstart guide](DOCKER_QUICKSTART.md))

```bash
./start_docker.sh
make -f Makefile.docker populate-db
# Open http://localhost:8080
```

---

## What's Running

### 1. PostgreSQL Database

- **Port**: 5432
- **Database**: `sstoryline`
- **User**: `sstoryline`
- **Password**: `sst_1234`

### 2. Web Server (API + UI)

- **Port**: 8080
- **URL**: http://localhost:8080
- **API Endpoint**: http://localhost:8080/searchN4L

### 3. CLI Tools (on-demand)

Command-line tools for loading data and running searches.

---

## Loading Data

### Load All Examples

```bash
make -f Makefile.docker populate-db
```

### Load a Single File

```bash
docker compose exec cli ./N4L examples/tutorial.n4l
```

### Load Your Own File

1. Copy your `.n4l` file to the `examples/` directory
2. Run:

```bash
docker compose exec cli ./N4L examples/yourfile.n4l
```

### Clear and Reload Database

```bash
docker compose exec cli sh -c "cd examples && ../N4L -u -wipe *.n4l"
```

---

## Managing Services

### Start

```bash
docker compose up -d
```

### Stop

```bash
docker compose down
```

### Restart

```bash
docker compose restart
```

### View Logs

```bash
# All services
docker compose logs -f

# Just the web server
docker compose logs -f server

# Just the database
docker compose logs -f postgres
```

### Check Status

```bash
docker compose ps
```

---

## Database Operations

### Access Database Shell

```bash
docker compose exec postgres psql -U sstoryline -d sstoryline
```

### Count Nodes in Database

```bash
docker compose exec postgres psql -U sstoryline -d sstoryline -c "SELECT COUNT(*) FROM Node;"
```

### Wipe Database and Start Fresh

```bash
docker compose down -v
docker compose up -d
make -f Makefile.docker populate-db
```

---

## CLI Tools

### Start CLI Container

```bash
docker compose --profile tools up -d cli
```

### Search

```bash
docker compose exec cli ./searchN4L "your search term"
```

### Other Tools Available

- `./N4L` - Load N4L files
- `./text2N4L` - Convert text to N4L
- `./searchN4L` - Search the database
- `./removeN4L` - Remove nodes
- `./pathsolve` - Path finding
- `./notes` - Notes management
- `./graph_report` - Generate reports

---

## Troubleshooting

### Check if services are running

```bash
docker compose ps
```

### Database not connecting

```bash
docker compose logs postgres
```

### Web server not responding

```bash
docker compose logs server
curl http://localhost:8080/status
```

### Rebuild after code changes

```bash
docker compose down
DOCKER_BUILDKIT=1 docker compose build
docker compose up -d
```

### Port already in use

If port 5432 or 8080 is already in use, edit `docker-compose.yml` to change the port mappings.

---

## Environment Variables

You can override database settings by editing `docker-compose.yml` or setting environment variables:

- `SST_DB_HOST` - Database host (default: postgres)
- `SST_DB_PORT` - Database port (default: 5432)
- `SST_DB_NAME` - Database name (default: sstoryline)
- `SST_DB_USER` - Database user (default: sstoryline)
- `SST_DB_PASSWORD` - Database password (default: sst_1234)

---

## Advanced Commands (using Makefile.docker)

```bash
make -f Makefile.docker help        # Show all available commands
make -f Makefile.docker build       # Build images
make -f Makefile.docker up          # Start services
make -f Makefile.docker down        # Stop services
make -f Makefile.docker logs        # View all logs
make -f Makefile.docker clean       # Remove everything
make -f Makefile.docker rebuild     # Rebuild and restart
make -f Makefile.docker status      # Container status
make -f Makefile.docker db-shell    # PostgreSQL shell
make -f Makefile.docker cli-shell   # CLI container shell
```
