Go gRPC Opentracing Instrumentation
===================================

[![GoDoc](https://godoc.org/github.com/charithe/otgrpc?status.svg)](https://godoc.org/github.com/charithe/otgrpc)

An attempt to use [Opentracing](http://opentracing.io/) with [gRPC](http://grpc.io) services. The official
[grpc-opentracing](https://github.com/grpc-ecosystem/grpc-opentracing) library currently only supports tracing
unary calls. This library makes use of gRPC `stats.Handler` interface to add tracing to gRPC streams as well.


The obvious approach to add tracing would be to make use of gRPC interceptors. However, the current interceptor 
interfaces lack the fuctionality to effectively add tracing information to the calls. Go gRPC has a built-in tracing 
mechanism that hooks into `x/net/trace`, but, the captured data is not accessible from external code. This leads us to
the `stats.Handler` interface which is primarily designed for stats gathering but, in the process, gives us access to 
all interesting events that happen during a RPC -- which can be used to gather the tracing information we need. 


Usage
-----

Grab the library:

```
go get github.com/charithe/otgrpc
```


Client side:

```go
tracer := // Tracer implementation

th := otgrpc.NewTraceHandler(tracer)
conn, err := grpc.Dial(address, grpc.WithStatsHandler(th))
...
```

Server side:

```go
tracer := // Tracer implementation

th := otgrpc.NewTraceHandler(tracer)
server := grpc.NewServer(grpc.StatsHandler(th))
...
```

### Options

Limit tracing to methods of your choosing

```go
tf := func(method string) bool {
    return method == "/my.svc/my.method"
}

th := otgrpc.NewTraceHandler(tracer, orgrpc.WithTraceEnabledFunc(tf))
```

Attach payloads as Span log events

```go
th := otgrpc.NewTraceHandler(tracer, otgrpc.WithPayloadLogging())
```

