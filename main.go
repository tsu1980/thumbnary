package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	d "runtime/debug"
	"strconv"
	"strings"
	"time"

	bimg "gopkg.in/h2non/bimg.v1"
)

var (
	aAddr                      = flag.String("a", "", "Bind address")
	aPort                      = flag.Int("p", 8088, "Port to listen")
	aVers                      = flag.Bool("v", false, "Show version")
	aVersl                     = flag.Bool("version", false, "Show version")
	aHelp                      = flag.Bool("h", false, "Show help")
	aHelpl                     = flag.Bool("help", false, "Show help")
	aCors                      = flag.Bool("cors", false, "Enable CORS support")
	aAuthForwarding            = flag.Bool("enable-auth-forwarding", false, "Forwards X-Forward-Authorization or Authorization header to the image source server. -enable-url-source flag must be defined. Tip: secure your server from public access to prevent attack vectors")
	aEnablePlaceholder         = flag.Bool("enable-placeholder", false, "Enable image response placeholder to be used in case of error")
	aOriginIdDetectMethods     = flag.String("origin-id-detect-methods", "header,query", "List of origin id detect methods(Comma separated)")
	aOriginIdDetectHostPattern = flag.String("origin-id-detect-host-pattern", "", "The regex pattern string for extract origin id from host name")
	aOriginIdDetectPathPattern = flag.String("origin-id-detect-path-pattern", "", "The regex pattern string for extract origin id from url path")
	aRedisURL                  = flag.String("redis-url", "", "The redis server url")
	aRedisChannelPrefix        = flag.String("redis-channel-prefix", "imarginary:", "The channel name for notification")
	aDBDriverName              = flag.String("db-driver-name", "", "The database driver name")
	aDBDataSourceName          = flag.String("db-data-source-name", "", "The database data source name")
	aMaxAllowedSize            = flag.Int("max-allowed-size", 0, "Restrict maximum size of http image source (in bytes)")
	aKey                       = flag.String("key", "", "Define API key for authorization")
	aCertFile                  = flag.String("certfile", "", "TLS certificate file path")
	aKeyFile                   = flag.String("keyfile", "", "TLS private key file path")
	aAuthorization             = flag.String("authorization", "", "Defines a constant Authorization header value passed to all the image source servers. -enable-url-source flag must be defined. This overwrites authorization headers forwarding behavior via X-Forward-Authorization")
	aPlaceholder               = flag.String("placeholder", "", "Image path to image custom placeholder to be used in case of error. Recommended minimum image size is: 1200x1200")
	aHTTPCacheTTL              = flag.Int("http-cache-ttl", -1, "The TTL in seconds")
	aReadTimeout               = flag.Int("http-read-timeout", 60, "HTTP read timeout in seconds")
	aWriteTimeout              = flag.Int("http-write-timeout", 60, "HTTP write timeout in seconds")
	aConcurrency               = flag.Int("concurrency", 0, "Throttle concurrency limit per second")
	aBurst                     = flag.Int("burst", 100, "Throttle burst max cache size")
	aMRelease                  = flag.Int("mrelease", 30, "OS memory release interval in seconds")
)

const usage = `thumbnary %s

Usage:
  thumbnary -p 80
  thumbnary -cors
  thumbnary -concurrency 10
  thumbnary -enable-auth-forwarding
  thumbnary -authorization "Basic AwDJdL2DbwrD=="
  thumbnary -enable-placeholder
  thumbnary -placeholder ./placeholder.jpg
  thumbnary -enable-url-signature -url-signature-key 4f46feebafc4b5e988f131c4ff8b5997 -url-signature-salt 88f131c4ff8b59974f46feebafc4b5e9
  thumbnary -h | -help
  thumbnary -v | -version

Options:
  -a <addr>                 Bind address [default: *]
  -p <port>                 Bind port [default: 8088]
  -h, -help                 Show help
  -v, -version              Show version
  -cors                     Enable CORS support [default: false]
  -key <key>                Define API key for authorization
  -http-cache-ttl <num>     The TTL in seconds. Adds caching headers to locally served files.
  -http-read-timeout <num>  HTTP read timeout in seconds [default: 30]
  -http-write-timeout <num> HTTP write timeout in seconds [default: 30]
  -enable-placeholder       Enable image response placeholder to be used in case of error [default: false]
  -enable-auth-forwarding   Forwards X-Forward-Authorization or Authorization header to the image source server. Tip: secure your server from public access to prevent attack vectors
  -enable-url-signature     Enable URL signature (URL-safe Base64-encoded HMAC digest) [default: false]
  -url-signature-key        The URL signature key (32 characters minimum)
  -url-signature-salt       The URL signature salt (32 characters minimum)
  -max-allowed-size <bytes> Restrict maximum size of http image source (in bytes)
  -certfile <path>          TLS certificate file path
  -keyfile <path>           TLS private key file path
  -authorization <value>    Defines a constant Authorization header value passed to all the image source servers. -enable-url-source flag must be defined. This overwrites authorization headers forwarding behavior via X-Forward-Authorization
  -placeholder <path>       Image path to image custom placeholder to be used in case of error. Recommended minimum image size is: 1200x1200
  -concurrency <num>        Throttle concurrency limit per second [default: disabled]
  -burst <num>              Throttle burst max cache size [default: 100]
  -mrelease <num>           OS memory release interval in seconds [default: 30]
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, Version))
	}
	flag.Parse()

	if *aHelp || *aHelpl {
		showUsage()
	}
	if *aVers || *aVersl {
		showVersion()
	}

	port := getPort(*aPort)

	opts := ServerOptions{
		Port:                      port,
		Address:                   *aAddr,
		CORS:                      *aCors,
		AuthForwarding:            *aAuthForwarding,
		EnablePlaceholder:         *aEnablePlaceholder,
		OriginIdDetectHostPattern: *aOriginIdDetectHostPattern,
		OriginIdDetectPathPattern: *aOriginIdDetectPathPattern,
		RedisURL:                  *aRedisURL,
		RedisChannelPrefix:        *aRedisChannelPrefix,
		DBDriverName:              *aDBDriverName,
		DBDataSourceName:          *aDBDataSourceName,
		APIKey:                    *aKey,
		Concurrency:               *aConcurrency,
		Burst:                     *aBurst,
		CertFile:                  *aCertFile,
		KeyFile:                   *aKeyFile,
		Placeholder:               *aPlaceholder,
		HTTPCacheTTL:              *aHTTPCacheTTL,
		HTTPReadTimeout:           *aReadTimeout,
		HTTPWriteTimeout:          *aWriteTimeout,
		Authorization:             *aAuthorization,
		MaxAllowedSize:            *aMaxAllowedSize,
	}

	// Create a memory release goroutine
	if *aMRelease > 0 {
		memoryRelease(*aMRelease)
	}

	// Validate HTTP cache param, if present
	if *aHTTPCacheTTL != -1 {
		checkHttpCacheTtl(*aHTTPCacheTTL)
	}

	// Parse origin id detect methods
	err := parseOriginIdDetectMethods(&opts, *aOriginIdDetectMethods)
	if err != nil {
		exitWithError(err.Error())
	}

	// Read placeholder image, if required
	if *aPlaceholder != "" {
		buf, err := ioutil.ReadFile(*aPlaceholder)
		if err != nil {
			exitWithError("cannot start the server: %s", err)
		}

		imageType := bimg.DetermineImageType(buf)
		if !bimg.IsImageTypeSupportedByVips(imageType).Load {
			exitWithError("Placeholder image type is not supported. Only JPEG, PNG or WEBP are supported")
		}

		opts.PlaceholderImage = buf
	} else if *aEnablePlaceholder {
		// Expose default placeholder
		opts.PlaceholderImage = placeholder
	}

	debug("thumbnary server listening on port :%d", opts.Port)

	// Load image source providers
	LoadSources(opts)

	// Create and open origin repository
	opts.OriginRepos, err = NewOriginRepository(OriginRepositoryTypeMySQL, opts)
	if err != nil {
		exitWithError("failed to create origin repository: %s", err.Error())
	}

	err = opts.OriginRepos.Open()
	if err != nil {
		exitWithError("failed to open origin repository: %s", err.Error())
	}

	// Start the server
	err = Server(opts)
	if err != nil {
		exitWithError("cannot start the server: %s", err)
	}
}

func getPort(port int) int {
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		newPort, _ := strconv.Atoi(portEnv)
		if newPort > 0 {
			port = newPort
		}
	}
	return port
}

func showUsage() {
	flag.Usage()
	os.Exit(1)
}

func showVersion() {
	fmt.Println(Version)
	os.Exit(1)
}

func checkHttpCacheTtl(ttl int) {
	if ttl < -1 || ttl > 31556926 {
		exitWithError("The -http-cache-ttl flag only accepts a value from 0 to 31556926")
	}

	if ttl == 0 {
		debug("Adding HTTP cache control headers set to prevent caching.")
	}
}

func parseOrigins(origins string) []*url.URL {
	urls := []*url.URL{}
	if origins == "" {
		return urls
	}
	for _, origin := range strings.Split(origins, ",") {
		u, err := url.Parse(origin)
		if err != nil {
			continue
		}
		urls = append(urls, u)
	}
	return urls
}

func parseOriginIdDetectMethods(o *ServerOptions, input string) error {
	methods := make([]OriginIdDetectMethod, 0, 5)
	for _, val := range strings.Split(input, ",") {
		val = strings.ToLower(strings.TrimSpace(val))
		_, ok := originIdDetectMethodMap[(OriginIdDetectMethod)(val)]
		if !ok {
			return fmt.Errorf("Unknown origin id detect method(%s)", val)
		}
		method := (OriginIdDetectMethod)(val)

		switch method {
		case OriginIdDetectMethod_Host:
			if o.OriginIdDetectHostPattern == "" {
				return fmt.Errorf("Missing required params: origin id detect host pattern")
			}
		case OriginIdDetectMethod_Path:
			if o.OriginIdDetectPathPattern == "" {
				return fmt.Errorf("Missing required params: origin id detect path pattern")
			}
		}

		methods = append(methods, method)
	}

	if len(methods) == 0 {
		return fmt.Errorf("origin id detect methods empty")
	}

	o.OriginIdDetectMethods = methods
	return nil
}

func memoryRelease(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		for range ticker.C {
			debug("FreeOSMemory()")
			d.FreeOSMemory()
		}
	}()
}

func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args)
	os.Exit(1)
}

func debug(msg string, values ...interface{}) {
	debug := os.Getenv("DEBUG")
	if debug == "thumbnary" || debug == "*" {
		log.Printf(msg, values...)
	}
}
