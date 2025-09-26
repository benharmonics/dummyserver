# dummyserver

A dummy HTTP server that accepts all requests sent to it

## Running

From the project directory, run the program with

```bash
go run main.go
```

By default, the server listens on port 9090. To change this, set the `-port`
flag e.g.:

```bash
go run main.go -port=8989
```

Run with the `-help` flag to see more options.

Send a request to the dummyserver e.g.

```bash
curl localhost:9090 --json '{ "message": "Hello World" }'
```

## Installing

You can install the executable with

```bash
go install
```

and then run it with the command `dummyserver`.

## Help

To get more information on the executable, run it with the `-help` flag:

```bash
dummyserver -help
```
