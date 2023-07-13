package session

import (
	"database/sql"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Session struct {
	CookieLifetime string
	CookiePersist  string
	CookieSecure   string
	CookieName     string
	CookieDomain   string
	SessionType    string
	DBPool         *sql.DB
	RedisPool      *redis.Pool
}

func (f *Session) InitSession() *scs.SessionManager {
	var persist, secure bool

	// how long should sessions last?

	minutes, err := strconv.Atoi(f.CookieLifetime)

	if err != nil {
		minutes = 60
	}

	// should cookies persist?

	if strings.ToLower(f.CookiePersist) == "true" {
		persist = true
	}

	// secure cookies?

	if strings.ToLower(f.CookieSecure) == "true" {
		secure = true
	}

	// create session

	session := scs.New()
	session.Lifetime = time.Duration(minutes) * time.Minute
	session.Cookie.Persist = persist
	session.Cookie.Name = f.CookieName
	session.Cookie.Secure = secure
	session.Cookie.Domain = f.CookieDomain
	session.Cookie.SameSite = http.SameSiteLaxMode

	// which session store?

	switch strings.ToLower(f.SessionType) {
	case "redis":
		session.Store = redisstore.New(f.RedisPool)
	case "mysql", "mariadb":
		session.Store = mysqlstore.New(f.DBPool)
	case "postgres", "postgresql":
		session.Store = postgresstore.New(f.DBPool)
	default:
	}

	return session
}
