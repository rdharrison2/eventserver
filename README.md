# eventserver

Test server to receive management events from MCU nodes.

## To build:

```
docker build . -t rdharrison2/event-server
```

## To run via docker:

```
docker run [-d] -p 8000:8000 --name event-server rdharrison2/event-server
```

And with TLS:

```
docker run [-d] -v $PWD:/etc/certs -p 8000:8000 --name event-server rdharrison2/event-server --use-tls
```

## To push:

```
docker push rdharrison2/event-server
```

## To run/test locally:

```
go run .
```

See --help for usage

```
go test . -v [-cover]
```
