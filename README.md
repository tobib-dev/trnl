# TRNL 

Trnl is a lightweight HTTP package written in Go.

## Motivation

This package was an opportunity to get into the TCP transport layer and experiment with the socket API provided by the Go net package. This app was also heavily inspired by the Go standard library net/http module, hence, the similarity between both modules.

Note: This is not meant to be a replacement for the net/http package or frameworks like Gin-Gonic or other production-ready libraries. In many ways, it was a way to see how far I can go with vanilla net package. Please feel free to test, extend, and use the module however you like.

## Installation

Trnl can be installed by running the command:
```
go get github.com/tobib-dev/trnl
```

Note: Installation of Go on your local machine is necessary to install this package

## Usage

Trnl can be used like standard HTTP packages are used. Here are some of the most used functions

### Generate Default Multiplexer
```
mux := trnl.Default()
```

### Initialize Server
```
srv = &trnl.Server{
  Addr: ":{PORT_NUMBER}",
  Handler: mux,
  ReadTimeout: 500 // Time Duration in milliseconds
}
```
