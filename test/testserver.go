package test

import (
	"io"
	"net"
	"time"

	"github.com/charithe/otgrpc"
	"github.com/opentracing/opentracing-go"

	"google.golang.org/grpc"

	context "golang.org/x/net/context"
)

type TestServer struct{}

func (ts *TestServer) UnaryRPC(ctx context.Context, req *TestRequest) (*TestResponse, error) {
	return &TestResponse{Response: req.Request}, nil
}

func (ts *TestServer) ClientStreamRPC(stream TestSvc_ClientStreamRPCServer) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&TestResponse{Response: "OK"})
		} else if err != nil {
			return err
		}
	}
}

func (ts *TestServer) ServerStreamRPC(req *TestRequest, stream TestSvc_ServerStreamRPCServer) error {
	for i := 0; i < 5; i++ {
		if err := stream.Send(&TestResponse{Response: req.Request}); err != nil {
			return err
		}
	}

	return nil
}

func (ts *TestServer) BidiStreamRPC(stream TestSvc_BidiStreamRPCServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			stream.Send(&TestResponse{Response: "OK"})
			return nil
		} else if err != nil {
			return err
		}

		if err = stream.Send(&TestResponse{Response: req.Request}); err != nil {
			return err
		}
	}
}

func StartInProcServer(socketPath string, tracer opentracing.Tracer, waitChan chan<- *grpc.Server) {
	lis, err := net.ListenUnix("unix", &net.UnixAddr{Name: socketPath, Net: "unix"})
	if err != nil {
		panic(err)
	}

	//server := grpc.NewServer(grpc.StreamInterceptor(otgrpc.StreamServerInterceptor(tracer)), grpc.UnaryInterceptor(otgrpc.UnaryServerInterceptor(tracer)))
	sh := otgrpc.NewTraceHandler(tracer, otgrpc.WithPayloadLogging())
	server := grpc.NewServer(grpc.StatsHandler(sh))
	RegisterTestSvcServer(server, &TestServer{})
	waitChan <- server
	if err = server.Serve(lis); err != nil {
		panic(err)
	}
}

func GetInProcClient(socketPath string, tracer opentracing.Tracer) TestSvcClient {
	dialFunc := func(addr string, timeout time.Duration) (net.Conn, error) {
		return net.DialUnix("unix", nil, &net.UnixAddr{Name: addr, Net: "unix"})
	}

	sh := otgrpc.NewTraceHandler(tracer, otgrpc.WithPayloadLogging())
	conn, err := grpc.Dial(socketPath,
		grpc.WithDialer(dialFunc),
		grpc.WithInsecure(),
		grpc.WithStatsHandler(sh))

	if err != nil {
		panic(err)
	}

	return NewTestSvcClient(conn)
}
