# sdrplay &ndash; A Golang wrapper of the SDRplay RSP API

sdrplay is a package that enables to use the RSP (by SDRplay) in a Go program. It uses GCO to wrap the SDRplay C library (version 1.97.1).

## Installation
In the code, the CGO is configured with this flags:
```
CFLAGS: -I/usr/local/include
LDFLAGS: -L/usr/local/lib -lmirsdrapi-rsp
```
So, to successfully compile this package, it is sufficient firstly install the API/Driver package released by SDRplay.
Then:
```
$ go get -u github.com/iclac/sdrplay
```
