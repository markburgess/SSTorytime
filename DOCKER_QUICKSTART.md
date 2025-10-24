# SSTorytime Docker - Quick Start

## First Time Setup (3 steps)

### 1. Build and Start

```bash
./start_docker.sh
```

This builds the Docker images and starts the database and web server.

### 2. Populate the Database

```bash
make -f Makefile.docker populate-db
```

This loads all example data into the database.

### 3. Open the Web Interface

```
http://localhost:8080
```

**That's it!** The system is now running.

---

## Common Operations

### Stop Everything

```bash
docker compose down
```

### Restart

```bash
docker compose up -d
```

### View Logs

```bash
docker compose logs -f server
```

### Add Your Own Data

```bash
# 1. Put your .n4l file in the examples/ directory
# 2. Load it:
docker compose exec cli ./N4L examples/yourfile.n4l
```

### Wipe Database and Start Fresh

```bash
docker compose down -v
./start_docker.sh
make -f Makefile.docker populate-db
```

---

See [DOCKER.md](DOCKER.md) for detailed documentation.
