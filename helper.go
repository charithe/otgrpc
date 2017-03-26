package otgrpc

import (
	"context"

	"google.golang.org/grpc/metadata"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	EventKey   = "event"
	PayloadKey = "payload"
)

var GRPCComponentTag = opentracing.Tag{Key: string(ext.Component), Value: "gRPC"}

func extractSpanContext(tracer opentracing.Tracer, ctx context.Context) opentracing.SpanContext {
	var sc opentracing.SpanContext
	sc = spanContextFromContext(ctx)
	if sc != nil {
		return sc
	}

	sc = extractSpanContextFromMetadata(tracer, ctx)
	return sc
}

func spanContextFromContext(ctx context.Context) opentracing.SpanContext {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		return parentSpan.Context()
	}
	return nil
}

func injectSpanToMetadata(tracer opentracing.Tracer, span opentracing.Span, ctx context.Context) (context.Context, error) {
	var md metadata.MD
	if tmpMD, ok := metadata.FromContext(ctx); ok {
		md = tmpMD.Copy()
	} else {
		md = metadata.New(nil)
	}

	if err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, NewMetadataReaderWriter(md)); err != nil {
		return ctx, err
	}

	return metadata.NewContext(ctx, md), nil
}

func extractSpanContextFromMetadata(tracer opentracing.Tracer, ctx context.Context) opentracing.SpanContext {
	var md metadata.MD
	if tmpMD, ok := metadata.FromContext(ctx); ok {
		md = tmpMD
	} else {
		md = metadata.New(nil)
	}

	// TODO How to deal with errors from Extract
	spanContext, _ := tracer.Extract(opentracing.HTTPHeaders, NewMetadataReaderWriter(md))
	return spanContext
}
