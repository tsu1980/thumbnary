package main

import (
    "regexp"
	"strings"
	"log"
	"fmt"
	"net"
	"net/http"
    "github.com/hashicorp/golang-lru"
    "github.com/gomodule/redigo/redis"
    "github.com/cenkalti/backoff"
	"database/sql"
    _ "github.com/go-sql-driver/mysql"
)

type OriginId string

type Origin struct {
    ID                          OriginId
	Scheme                      string
	Host                        string
	PathPrefix                  string
	URLSignatureKey             string
	URLSignatureKey_Previous    string
	URLSignatureKey_Version     string
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
        o.RedisChannelPrefix + "notification:origin_updated",
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
    if strings.Contains(msg.Channel, "notification:origin_updated") {
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

func FetchOriginRecord(originId OriginId) (*Origin, error) {
    cval, ok := originCache.Get(originId)
    if ok {
        if origin, ok := cval.(*Origin); ok {
            return origin, nil
        }
    }

    origin := &Origin{}
    sql := "SELECT ID, Scheme, Host, PathPrefix, URLSignatureKey, URLSignatureKey_Previous, URLSignatureKey_Version FROM origin WHERE ID = ?"
    err := db.QueryRow(sql, (string)(originId)).Scan(
        &origin.ID,
        &origin.Scheme,
        &origin.Host,
        &origin.PathPrefix,
        &origin.URLSignatureKey,
        &origin.URLSignatureKey_Previous,
        &origin.URLSignatureKey_Version,
    )
    if err != nil {
		return nil, fmt.Errorf("Cannot select origin id: (originId=%s) (err=%v)", originId, err)
    }
    log.Printf("Origin[%s] fetched\n", originId)
    originCache.Add(originId, origin)
    return origin, nil
}

func FindOrigin(o ServerOptions, req *http.Request) (*Origin, error) {
    host, _, _ := net.SplitHostPort(req.Host)

    r := regexp.MustCompile(o.OriginHostPattern)
    group := r.FindStringSubmatch(host)
    if group == nil {
		return nil, fmt.Errorf("Cannot extract origin id: (host=%s)", host)
    }

    var originId = OriginId(group[1]);
    if originId == "" {
    	return nil, fmt.Errorf("Origin id is empty: (host=%s)", host)
    }

    origin, err := FetchOriginRecord(originId)
    if err != nil {
    	return nil, err
    }
	return origin, nil
}
