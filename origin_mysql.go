package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/cenkalti/backoff"
	"github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/hashicorp/golang-lru"
)

const OriginRepositoryTypeMySQL OriginRepositoryType = "mysql"

type MySQLOriginRepository struct {
	Options ServerOptions
}

func NewMySQLOriginRepository(o ServerOptions) OriginRepository {
	return &MySQLOriginRepository{Options: o}
}

func (repo *MySQLOriginRepository) Open() error {
	err := OpenDB(repo.Options)
	if err != nil {
		return fmt.Errorf("failed to open database: %s", err)
	}

	err = StartRedis(repo.Options)
	if err != nil {
		return fmt.Errorf("failed to start redis: %s", err)
	}

	return nil
}

func (repo *MySQLOriginRepository) Close() {
}

var db *sql.DB
var originCache *lru.ARCCache

// OpenDB open database connection
func OpenDB(o ServerOptions) error {
	var err = (error)(nil)

	if o.DBTlsKeyName != "" && o.DBTlsServerHostName != "" && o.DBTlsServerCAPem != "" && o.DBTlsClientCertPem != "" && o.DBTlsClientKeyPem != "" {
		serverCAPem, err := base64.StdEncoding.DecodeString(o.DBTlsServerCAPem)
		if err != nil {
			return fmt.Errorf("Failed to decode server ca PEM value")
		}
		clientCertPem, err := base64.StdEncoding.DecodeString(o.DBTlsClientCertPem)
		if err != nil {
			return fmt.Errorf("Failed to decode client certificate PEM value")
		}
		clientKeyPem, err := base64.StdEncoding.DecodeString(o.DBTlsClientKeyPem)
		if err != nil {
			return fmt.Errorf("Failed to decode client key PEM value")
		}

		rootCertPool := x509.NewCertPool()
		if ok := rootCertPool.AppendCertsFromPEM(serverCAPem); !ok {
			return fmt.Errorf("Failed to append server CA PEM")
		}

		clientCert := make([]tls.Certificate, 0, 1)
		certs, err := tls.X509KeyPair(clientCertPem, clientKeyPem)
		if err != nil {
			return err
		}

		clientCert = append(clientCert, certs)
		err = mysql.RegisterTLSConfig(o.DBTlsKeyName, &tls.Config{
			RootCAs:      rootCertPool,
			Certificates: clientCert,
			ServerName:   o.DBTlsServerHostName,
		})
		if err != nil {
			return err
		}
	}

	db, err = sql.Open(o.DBDriverName, o.DBDataSourceName)
	if err != nil {
		return err
	}

	// Initialize memory cache
	originCache, err = lru.NewARC(10)
	if err != nil {
		return err
	}

	return nil
}

var bo *backoff.ExponentialBackOff

func StartRedis(o ServerOptions) error {
	// Start a goroutine to receive notifications.
	go func() {
		operation := func() error {
			return ListenNotifications(o)
		}

		bo = backoff.NewExponentialBackOff()
		err := backoff.Retry(operation, bo)
		if err != nil {
			log.Printf("Cannot listen notifications (err=%s)", err.Error())
			panic("fatal error")
		}
		log.Printf("Listen notifications finished")
	}()

	return nil
}

func ListenNotifications(o ServerOptions) error {
	conn, err := redis.DialURL(o.RedisURL)
	if err != nil {
		log.Printf("failed to connect to redis (url=%s)\n", o.RedisURL)
		return err
	}
	defer conn.Close()

	psc := redis.PubSubConn{Conn: conn}

	channels := []string{
		o.RedisChannelPrefix + "origin:changed",
	}
	err = psc.Subscribe(redis.Args{}.AddFlat(channels)...)
	if err != nil {
		log.Printf("failed to subscribe redis channels (url=%s)", o.RedisURL)
		return err
	}

	bo.Reset()
	clearOriginsCacheAll()

	for conn.Err() == nil {
		switch msg := psc.Receive().(type) {
		case redis.Message:
			onMessage(msg)
		case redis.Subscription:
			log.Printf("%s: %s %d\n", msg.Channel, msg.Kind, msg.Count)
		case error:
			log.Printf("[ListenNotifications] Error occured: %s\n", msg.Error())
			return msg
		}
	}

	return nil
}

func onMessage(msg redis.Message) {
	if strings.Contains(msg.Channel, "origin:changed") {
		onOriginUpdated((OriginSlug)(string(msg.Data)))
	} else {
		log.Printf("%s: Unknown message: %s\n", msg.Channel, msg.Data)
	}
}

func onOriginUpdated(originSlug OriginSlug) {
	log.Printf("Origin[%s] updated\n", originSlug)
	originCache.Remove(originSlug)
}

func clearOriginsCacheAll() {
	originCache.Purge()
	log.Printf("Origins cache cleared")
}

func (repo *MySQLOriginRepository) Get(originSlug OriginSlug) (*Origin, error) {
	cval, ok := originCache.Get(originSlug)
	if ok {
		if origin, ok := cval.(*Origin); ok {
			return origin, nil
		}
	}

	origin := &Origin{}
	sql := fmt.Sprintf("SELECT Slug, SourceType, Scheme, Host, PathPrefix, URLSignatureEnabled, URLSignatureKey, URLSignatureKey_Previous, URLSignatureKey_Version, AllowExternalHTTPSource FROM %s WHERE Slug = ?",
		repo.Options.OriginTableName)
	err := db.QueryRow(sql, (string)(originSlug)).Scan(
		&origin.Slug,
		&origin.SourceType,
		&origin.Scheme,
		&origin.Host,
		&origin.PathPrefix,
		&origin.URLSignatureEnabled,
		&origin.URLSignatureKey,
		&origin.URLSignatureKey_Previous,
		&origin.URLSignatureKey_Version,
		&origin.AllowExternalHTTPSource,
	)
	if err != nil {
		return nil, fmt.Errorf("Cannot select origin slug: (originSlug=%s) (err=%v)", originSlug, err)
	}
	log.Printf("Origin[%s] fetched (origin=%+v)\n", originSlug, origin)
	originCache.Add(originSlug, origin)
	return origin, nil
}
