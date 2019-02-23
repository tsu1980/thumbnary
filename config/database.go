package config

type DatabaseConfiguration struct {
	// The redis server url
	RedisURL string

	// The channel name for notification
	RedisChannelPrefix string

	// The database driver name
	DBDriverName string

	// The database data source name
	DBDataSourceName string

	// The key name for TLS database connection
	DBTlsKeyName string

	// The server hostname for TLS database connection
	DBTlsServerHostName string

	// Base64 encoded server CA PEM for TLS database connection
	DBTlsServerCAPem string

	// Base64 encoded client certificate PEM for TLS database connection
	DBTlsClientCertPem string

	// Base64 encoded client key PEM for TLS database connection
	DBTlsClientKeyPem string

	// The database origin table name
	OriginTableName string
}
