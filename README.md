# dummyserver

A dummy HTTP server that accepts all requests sent to it

## Running

From the project directory, run the program with

```bash
go run main.go
```

Alternatively, you can build the executable with

```bash
go build
```

and then run it with the command `dummyserver`.

To listen on a specific port, set the `-port` flag, i.e.

```bash
dummyserver -port 5432
```

## Help

To get more information on the executable, run it with the `-help` flag:

```bash
dummyserver -help
```
