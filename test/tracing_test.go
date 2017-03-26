package test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"google.golang.org/grpc"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

var (
	client TestSvcClient
	tracer *mocktracer.MockTracer
)

func TestMain(m *testing.M) {
	socketPath := "/tmp/testsvc"
	defer os.Remove(socketPath)

	tracer = mocktracer.New()
	waitChan := make(chan *grpc.Server)
	go StartInProcServer(socketPath, tracer, waitChan)
	server := <-waitChan

	client = GetInProcClient(socketPath, tracer)
	retVal := m.Run()
	server.Stop()
	os.Exit(retVal)
}

func runTest(t *testing.T, body func(context.Context)) {
	tracer.Reset()

	ctx, parentSpan := createContext()
	body(ctx)
	parentSpan.Finish()

	spans := tracer.FinishedSpans()
	assert.NotEmpty(t, spans)
	//assert.True(t, len(spans) > 1)

	var prevParentSpanID int
	for _, s := range spans {
		printSpan(s)
		if prevParentSpanID != 0 {
			assert.Equal(t, prevParentSpanID, s.SpanContext.SpanID)
		}
		prevParentSpanID = s.ParentID
		assert.Equal(t, parentSpan.SpanContext.TraceID, s.SpanContext.TraceID)
	}
}

func TestUnary(t *testing.T) {
	runTest(t, func(ctx context.Context) {
		res, err := client.UnaryRPC(ctx, request("ping"))
		assert.NoError(t, err)
		assert.Equal(t, "ping", res.Response)
	})
}

func TestClientStream(t *testing.T) {
	runTest(t, func(ctx context.Context) {
		stream, err := client.ClientStreamRPC(ctx)
		assert.NoError(t, err)
		for i := 0; i < 5; i++ {
			err = stream.Send(request(fmt.Sprintf("ping%d", i)))
			assert.NoError(t, err)
		}

		resp, err := stream.CloseAndRecv()
		assert.NoError(t, err)
		assert.Equal(t, "OK", resp.Response)
	})
}

func TestServerStream(t *testing.T) {
	runTest(t, func(ctx context.Context) {
		stream, err := client.ServerStreamRPC(ctx, request("ping"))
		assert.NoError(t, err)
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			assert.NoError(t, err)
			assert.Equal(t, "ping", resp.Response)
		}
	})
}

func TestBidiStream(t *testing.T) {
	runTest(t, func(ctx context.Context) {
		stream, err := client.BidiStreamRPC(ctx)
		assert.NoError(t, err)
		for i := 0; i < 5; i++ {
			msg := fmt.Sprintf("ping%d", i)
			err = stream.Send(request(msg))
			assert.NoError(t, err)
			resp, err := stream.Recv()
			assert.NoError(t, err)
			assert.Equal(t, msg, resp.Response)
		}
		err = stream.CloseSend()
		assert.NoError(t, err)

		for {
			_, err = stream.Recv()
			if err == io.EOF {
				break
			}
			assert.NoError(t, err)
		}
	})
}

func createContext() (context.Context, *mocktracer.MockSpan) {
	span := tracer.StartSpan("grpc_op")
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	mockSpan := span.(*mocktracer.MockSpan)
	return ctx, mockSpan
}

func request(body string) *TestRequest {
	return &TestRequest{Request: body}
}

func printSpan(span *mocktracer.MockSpan) {
	fmt.Printf("%s\n", span)
	fmt.Printf("\t[%+v]\n", span.Tags())
	for _, lr := range span.Logs() {
		for _, f := range lr.Fields {
			fmt.Printf("\t%s %+v\n", lr.Timestamp, f)
		}
	}
	fmt.Printf("\n\n")
}
