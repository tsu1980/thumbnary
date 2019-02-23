package main

import (
	"encoding/json"
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

	"github.com/spf13/viper"
	"github.com/tsu1980/thumbnary/config"
	bimg "gopkg.in/h2non/bimg.v1"
)

var (
	aConfigFile = flag.String("config", "", "Config file path")
	aVers       = flag.Bool("v", false, "Show version")
	aVersl      = flag.Bool("version", false, "Show version")
	aHelp       = flag.Bool("h", false, "Show help")
	aHelpl      = flag.Bool("help", false, "Show help")
)

const usage = `thumbnary %s

Usage:
  thumbnary -config ./config.yml
  thumbnary -h | -help
  thumbnary -v | -version

Options:
  -config <path>            Config file path
  -h, -help                 Show help
  -v, -version              Show version
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

	var config config.Configuration
	viper.SetDefault("Server.Port", "8080")
	viper.SetDefault("Server.Cors", false)
	viper.SetDefault("Server.AuthForwarding", false)
	viper.SetDefault("Server.EnablePlaceholder", false)
	viper.SetDefault("Server.OriginSlugDetectMethods", "header,query")
	viper.SetDefault("Server.OriginSlugDetectHostPattern", "")
	viper.SetDefault("Server.OriginSlugDetectPathPattern", "")
	viper.SetDefault("Server.MaxAllowedSize", 0)
	viper.SetDefault("Server.MaxOutputMP", 0)
	viper.SetDefault("Server.HTTPCacheTTL", -1)
	viper.SetDefault("Server.ReadTimeout", 60)
	viper.SetDefault("Server.WriteTimeout", 60)
	viper.SetDefault("Server.Concurrency", 0)
	viper.SetDefault("Server.Burst", 100)
	viper.SetDefault("Server.MRelease", 30)
	viper.SetDefault("Database.RedisURL", "")
	viper.SetDefault("Database.RedisChannelPrefix", "thumbnary:")
	viper.SetDefault("Database.DBDriverName", "mysql")
	viper.SetDefault("Database.DBDataSourceName", "")
	viper.SetDefault("Database.DBTlsKeyName", "")
	viper.SetDefault("Database.DBTlsServerHostName", "")
	viper.SetDefault("Database.DBTlsServerCAPem", "")
	viper.SetDefault("Database.DBTlsClientCertPem", "")
	viper.SetDefault("Database.DBTlsClientKeyPem", "")
	viper.SetDefault("Database.OriginTableName", "origin")

	if *aConfigFile != "" {
		viper.SetConfigFile(*aConfigFile)
		if err := viper.ReadInConfig(); err != nil {
			exitWithError("Error reading config file, %s", err)
		}
	}
	err := viper.Unmarshal(&config)
	if err != nil {
		exitWithError("unable to decode into struct, %v", err)
	}

	bs, err := json.Marshal(&config)
	if err != nil {
		exitWithError("unable to marshal config to JSON: %v", err)
	}
	log.Printf("config = %s", string(bs))

	port := getPort(config.Server.Port)

	opts := ServerOptions{
		Port:                        port,
		Address:                     config.Server.Addr,
		CORS:                        config.Server.Cors,
		AuthForwarding:              config.Server.AuthForwarding,
		EnablePlaceholder:           config.Server.EnablePlaceholder,
		OriginSlugDetectHostPattern: config.Server.OriginSlugDetectHostPattern,
		OriginSlugDetectPathPattern: config.Server.OriginSlugDetectPathPattern,
		RedisURL:                    config.Database.RedisURL,
		RedisChannelPrefix:          config.Database.RedisChannelPrefix,
		DBDriverName:                config.Database.DBDriverName,
		DBDataSourceName:            config.Database.DBDataSourceName,
		DBTlsKeyName:                config.Database.DBTlsKeyName,
		DBTlsServerHostName:         config.Database.DBTlsServerHostName,
		DBTlsServerCAPem:            config.Database.DBTlsServerCAPem,
		DBTlsClientCertPem:          config.Database.DBTlsClientCertPem,
		DBTlsClientKeyPem:           config.Database.DBTlsClientKeyPem,
		OriginTableName:             config.Database.OriginTableName,
		APIKey:                      config.Server.Key,
		Concurrency:                 config.Server.Concurrency,
		Burst:                       config.Server.Burst,
		CertFile:                    config.Server.CertFile,
		KeyFile:                     config.Server.KeyFile,
		Placeholder:                 config.Server.Placeholder,
		HTTPCacheTTL:                config.Server.HTTPCacheTTL,
		HTTPReadTimeout:             config.Server.ReadTimeout,
		HTTPWriteTimeout:            config.Server.WriteTimeout,
		Authorization:               config.Server.Authorization,
		MaxAllowedSize:              config.Server.MaxAllowedSize,
		MaxOutputMP:                 config.Server.MaxOutputMP,
	}

	// Create a memory release goroutine
	if config.Server.MRelease > 0 {
		memoryRelease(config.Server.MRelease)
	}

	// Validate HTTP cache param, if present
	if config.Server.HTTPCacheTTL != -1 {
		checkHttpCacheTtl(config.Server.HTTPCacheTTL)
	}

	// Parse origin slug detect methods
	err = parseOriginSlugDetectMethods(&opts, config.Server.OriginSlugDetectMethods)
	if err != nil {
		exitWithError(err.Error())
	}

	// Read placeholder image, if required
	if config.Server.Placeholder != "" {
		buf, err := ioutil.ReadFile(config.Server.Placeholder)
		if err != nil {
			exitWithError("cannot start the server: %s", err)
		}

		imageType := bimg.DetermineImageType(buf)
		if !bimg.IsImageTypeSupportedByVips(imageType).Load {
			exitWithError("Placeholder image type is not supported. Only JPEG, PNG or WEBP are supported")
		}

		opts.PlaceholderImage = buf
	} else if config.Server.EnablePlaceholder {
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
	Server(opts)
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
	os.Exit(0)
}

func showVersion() {
	fmt.Println(Version)
	os.Exit(0)
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

func parseOriginSlugDetectMethods(o *ServerOptions, input string) error {
	methods := make([]OriginSlugDetectMethod, 0, 5)
	for _, val := range strings.Split(input, ",") {
		val = strings.ToLower(strings.TrimSpace(val))
		_, ok := originSlugDetectMethodMap[(OriginSlugDetectMethod)(val)]
		if !ok {
			return fmt.Errorf("Unknown origin slug detect method(%s)", val)
		}
		method := (OriginSlugDetectMethod)(val)

		switch method {
		case OriginSlugDetectMethod_Host:
			if o.OriginSlugDetectHostPattern == "" {
				return fmt.Errorf("Missing required params: origin slug detect host pattern")
			}
		case OriginSlugDetectMethod_Path:
			if o.OriginSlugDetectPathPattern == "" {
				return fmt.Errorf("Missing required params: origin slug detect path pattern")
			}
		}

		methods = append(methods, method)
	}

	if len(methods) == 0 {
		return fmt.Errorf("origin slug detect methods empty")
	}

	o.OriginSlugDetectMethods = methods
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
