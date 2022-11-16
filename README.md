# tcpfw

tcpfw is a tcp connection forwarder, plain and simple.

## Install

With Go installed, install via:

```shell
go install github.com/kernle32dll/tcpfw@latest
```

## How to use

| ARGUMENT  | DESCRIPTION                              |
|-----------|------------------------------------------|
| --help    | displays a help                          |
| --host    | host to listen to, defaults to localhost |
| --inport  | port to listen to                        |
| --outport | port to forward the connection to        |