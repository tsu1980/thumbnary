package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/cenkalti/backoff"
	_ "github.com/go-sql-driver/mysql"
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
		fmt.Errorf("failed to open database: %s", err)
	}

	err = StartRedis(repo.Options)
	if err != nil {
		fmt.Errorf("failed to start redis: %s", err)
	}

	return nil
}

func (repo *MySQLOriginRepository) Close() {
}

var db *sql.DB
var originCache *lru.ARCCache

func OpenDB(o ServerOptions) error {
	var err = (error)(nil)
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
		onOriginUpdated((OriginId)(string(msg.Data)))
	} else {
		log.Printf("%s: Unknown message: %s\n", msg.Channel, msg.Data)
	}
}

func onOriginUpdated(originId OriginId) {
	log.Printf("Origin[%s] updated\n", originId)
	originCache.Remove(originId)
}

func clearOriginsCacheAll() {
	originCache.Purge()
	log.Printf("Origins cache cleared")
}

func (repo *MySQLOriginRepository) Get(originId OriginId) (*Origin, error) {
	cval, ok := originCache.Get(originId)
	if ok {
		if origin, ok := cval.(*Origin); ok {
			return origin, nil
		}
	}

	origin := &Origin{}
	sql := fmt.Sprintf("SELECT ID, Scheme, Host, PathPrefix, URLSignatureKeyEnabled, URLSignatureKey, URLSignatureKey_Previous, URLSignatureKey_Version FROM %s WHERE ID = ?",
		repo.Options.DBOriginTableName)
	err := db.QueryRow(sql, (string)(originId)).Scan(
		&origin.ID,
		&origin.SourceType,
		&origin.Scheme,
		&origin.Host,
		&origin.PathPrefix,
		&origin.URLSignatureEnabled,
		&origin.URLSignatureKey,
		&origin.URLSignatureKey_Previous,
		&origin.URLSignatureKey_Version,
	)
	if err != nil {
		return nil, fmt.Errorf("Cannot select origin id: (originId=%s) (err=%v)", originId, err)
	}
	log.Printf("Origin[%s] fetched (origin=%+v)\n", originId, origin)
	originCache.Add(originId, origin)
	return origin, nil
}
