package otgrpc

type Option func(*options)

type options struct {
	traceEnabledFunc func(method string) bool
	logPayloads      bool
}

func newOptions(opts ...Option) *options {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	if o.traceEnabledFunc == nil {
		o.traceEnabledFunc = func(string) bool { return true }
	}
	return o
}

// WithTraceEnabledFunc defines a function that indicates to the tracing implementation whether the method should be traced or not
func WithTraceEnabledFunc(f func(method string) bool) Option {
	return func(opt *options) {
		opt.traceEnabledFunc = f
	}
}

//WithPayloadLogging enables logging of RPC payloads
func WithPayloadLogging() Option {
	return func(opt *options) {
		opt.logPayloads = true
	}
}
