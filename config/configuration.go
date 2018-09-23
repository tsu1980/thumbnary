package config

type Configuration struct {
	Server ServerConfiguration

	// The redis server url
	RedisURL string

	// The channel name for notification
	RedisChannelPrefix string

	// The database driver name
	DBDriverName string

	// The database data source name
	DBDataSourceName string

	// The database origin table name
	DBOriginTableName string
}
