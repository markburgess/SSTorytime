## Running the SSTorytime database in docker

Running the postgreSQL database server dependency to SSTorytime in a docker container
makes is easy once the docker daemon and the docker-compose cli is
installed. The docker-compose cli is installed automaticallt when instolling
docker desktop.

For installation instructions, see https://docs.docker.com/desktop/

The following command will start a postgreSQL container with all the necessary
configuration to use it with the SSTorytime cli tools, and the http-server that
visualizes searches in the database.

```
$ docker-compose up # Use -d option if it should run in the background
```

Then the postgreSQL server should be available on `127.0.0.1:5432` with the db
name, user and password that the SSTorytime utilities expect.

To stop the postgreSQL server:

```
$ docker compose down # When started in daemon mode with -d. Otherwise just Ctrl^c in the terminal window 
```

## Context

The setup creates a docker volume which serves as persistent storage for the
database server. The volume persists with it's data until it is deleted.

To inspect it:
```
$ docker volume ls
docker volume ls
DRIVER    VOLUME NAME
local     docker_postgres_data
$ docker inspect volume docker_postgres_data
 docker inspect  docker_postgres_data
[
    {
        "CreatedAt": "2025-11-05T16:04:37Z",
        "Driver": "local",
        "Labels": {
            "com.docker.compose.config-hash": "fc9c3243864dd27a923dba18ebe0743bc1bd63150910bb22573d03430e222307",
            "com.docker.compose.project": "docker",
            "com.docker.compose.version": "2.40.3",
            "com.docker.compose.volume": "postgres_data"
        },
        "Mountpoint": "/var/lib/docker/volumes/docker_postgres_data/_data",
        "Name": "docker_postgres_data",
        "Options": null,
        "Scope": "local"
    }
]
```

It can be deleted like this (when container is removed first): 
```
# Remove stopped container
$ docker rm sstorytime-postgres
sstorytime-postgres
# Remove volume
$ docker volume rm docker_postgres_data
docker_postgres_data
```

The volume will be recreated (empty of course) next time `docker-compose up`is run.
