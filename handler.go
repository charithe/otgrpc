package otgrpc

//Provides Opentracing instrumentation for gRPC services

import (
	"fmt"

	context "golang.org/x/net/context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/stats"
)

type TraceHandler struct {
	tracer opentracing.Tracer
	opts   *options
}

// NewTraceHandler creates a gRPC stats.Handler instance that instruments RPCs with Opentracing trace contexts
func NewTraceHandler(tracer opentracing.Tracer, o ...Option) *TraceHandler {
	return &TraceHandler{
		tracer: tracer,
		opts:   newOptions(o...),
	}
}

func (th *TraceHandler) TagRPC(ctx context.Context, tagInfo *stats.RPCTagInfo) context.Context {
	if !th.opts.traceEnabledFunc(tagInfo.FullMethodName) {
		return ctx
	}

	spanCtx := extractSpanContext(th.tracer, ctx)
	span := th.tracer.StartSpan(tagInfo.FullMethodName, opentracing.FollowsFrom(spanCtx), GRPCComponentTag)
	newCtx, _ := injectSpanToMetadata(th.tracer, span, ctx)
	return opentracing.ContextWithSpan(newCtx, span)
}

func (th *TraceHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	switch t := s.(type) {
	case *stats.Begin:
		span.LogFields(log.String(EventKey, "RPC started"))
	case *stats.InPayload:
		e := log.String(EventKey, fmt.Sprintf("Payload received: Wire length=%d", t.WireLength))
		if th.opts.logPayloads {
			span.LogFields(e, log.Object(PayloadKey, t.Payload))
		} else {
			span.LogFields(e)
		}
	case *stats.InHeader:
		span.LogFields(log.String(EventKey, fmt.Sprintf("Header received: Remote addr=%s, Local addr=%s", t.RemoteAddr, t.LocalAddr)))
	case *stats.InTrailer:
		span.LogFields(log.String(EventKey, "Trailer received"))
	case *stats.OutPayload:
		e := log.String(EventKey, fmt.Sprintf("Payload sent: Wire length=%d", t.WireLength))
		if th.opts.logPayloads {
			span.LogFields(e, log.Object(PayloadKey, t.Payload))
		} else {
			span.LogFields(e)
		}
	case *stats.OutHeader:
		span.LogFields(log.String(EventKey, fmt.Sprintf("Header sent: Remote addr=%s, Local addr=%s", t.RemoteAddr, t.LocalAddr)))
	case *stats.OutTrailer:
		span.LogFields(log.String(EventKey, "Trailer sent"))
	case *stats.End:
		if t.IsClient() {
			span.SetTag(string(ext.SpanKind), ext.SpanKindRPCClientEnum)
		} else {
			span.SetTag(string(ext.SpanKind), ext.SpanKindRPCServerEnum)
		}

		if t.Error != nil {
			span.SetTag(string(ext.Error), true)
			span.LogFields(log.String(EventKey, "RPC failed"), log.Error(t.Error))
		} else {
			span.LogFields(log.String(EventKey, "RPC ended"))
		}
		span.Finish()
	}
}

func (th *TraceHandler) TagConn(ctx context.Context, tagInfo *stats.ConnTagInfo) context.Context {
	return ctx
}

func (th *TraceHandler) HandleConn(ctx context.Context, s stats.ConnStats) {}
