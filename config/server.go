package config

type ServerConfiguration struct {
	// Bind address
	Addr string

	// Port to listen
	Port int

	// Enable CORS support
	Cors bool

	// Forwards X-Forward-Authorization or Authorization header to the image source server.
	// -enable-url-source flag must be defined.
	// Tip: secure your server from public access to prevent attack vectors
	AuthForwarding bool

	// Enable image response placeholder to be used in case of error
	EnablePlaceholder bool

	// List of origin slug detect methods(Comma separated)
	// host, path, query, header, urlsig are allowed
	OriginSlugDetectMethods string

	// The regex pattern string for extract origin slug from host name
	OriginSlugDetectHostPattern string

	// The regex pattern string for extract origin slug from url path
	OriginSlugDetectPathPattern string

	// Restrict maximum size of http image source (in bytes)
	MaxAllowedSize int

	// Define API key for authorization
	Key string

	// TLS certificate file path
	CertFile string

	// TLS private key file path
	KeyFile string

	// Defines a constant Authorization header value passed to all the image source servers.
	// -enable-url-source flag must be defined.
	// This overwrites authorization headers forwarding behavior via X-Forward-Authorization
	Authorization string

	// Image path to image custom placeholder to be used in case of error.
	// Recommended minimum image size is: 1200x1200
	Placeholder string

	// The TTL in seconds
	HTTPCacheTTL int

	// HTTP read timeout in seconds
	ReadTimeout int

	// HTTP write timeout in seconds
	WriteTimeout int

	// Throttle concurrency limit per second
	Concurrency int

	// Throttle burst max cache size
	Burst int

	// OS memory release interval in seconds
	MRelease int
}
