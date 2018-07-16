package datadog

type options struct {
	noCompression bool
}

// Option represents client option.
type Option func(*options)

func applyOptions(o *options, opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// ----------------------------------------------------------------------------

// WithCompressionDisabled disables HTTP request compression.
func WithCompressionDisabled() Option {
	return func(o *options) {
		o.noCompression = true
	}
}
