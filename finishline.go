package finishline

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/ericfisherdev/gofresh-api/cache"
	"github.com/ericfisherdev/gofresh-api/mailer"
	"github.com/ericfisherdev/gofresh-api/render"
	"github.com/ericfisherdev/gofresh-api/session"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const version = "1.0.0"

var myRedisCache *cache.RedisCache
var myBadgerCache *cache.BadgerCache
var redisPool *redis.Pool
var badgerConn *badger.DB

type FinishLine struct {
	AppName       string
	Debug         bool
	Version       string
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	RootPath      string
	Routes        *chi.Mux
	Render        *render.Render
	Session       *scs.SessionManager
	DB            Database
	JetViews      *jet.Set
	config        config
	EncryptionKey string
	Cache         cache.Cache
	Scheduler     *cron.Cron
	Mail          mailer.Mail
	Server        Server
}

type Server struct {
	ServerName string
	Port       string
	Secure     bool
	URL        string
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
	redis       redisConfig
}

func (f *FinishLine) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "mail", "data", "public", "tmp", "logs", "middleware"},
	}

	err := f.Init(pathConfig)
	if err != nil {
		return err
	}

	err = f.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	// read .env
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	// create loggers
	infoLog, errorLog := f.startLoggers()

	// connect to database
	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := f.OpenDB(os.Getenv("DATABASE_TYPE"), f.BuildDSN())
		if err != nil {
			errorLog.Println(err)
			os.Exit(1)
		}

		f.DB = Database{
			DataType: os.Getenv("DATABASE_TYPE"),
			Pool:     db,
		}
	}

	scheduler := cron.New()
	f.Scheduler = scheduler

	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		myRedisCache = f.createClientRedisCache()
		f.Cache = myRedisCache
		redisPool = myRedisCache.Conn
	}

	if os.Getenv("CACHE") == "badger" {
		myBadgerCache = f.createClientBadgerCache()
		f.Cache = myBadgerCache
		badgerConn = myBadgerCache.Conn

		_, err = f.Scheduler.AddFunc("@daily", func() {
			_ = myBadgerCache.Conn.RunValueLogGC(0.7)
		})

		if err != nil {
			return err
		}
	}

	f.InfoLog = infoLog
	f.ErrorLog = errorLog
	f.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	f.Version = version
	f.RootPath = rootPath
	f.Mail = f.createMailer()
	f.Routes = f.routes().(*chi.Mux)

	f.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      f.BuildDSN(),
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	secure := true
	if strings.ToLower(os.Getenv("SECURE")) == "false" {
		secure = false
	}

	f.Server = Server{
		ServerName: os.Getenv("SERVER_NAME"),
		Port:       os.Getenv("PORT"),
		Secure:     secure,
		URL:        os.Getenv("APP_URL"),
	}

	// create session

	sess := session.Session{
		CookieLifetime: f.config.cookie.lifetime,
		CookiePersist:  f.config.cookie.persist,
		CookieName:     f.config.cookie.name,
		SessionType:    f.config.sessionType,
		CookieDomain:   f.config.cookie.domain,
	}

	switch f.config.sessionType {
	case "redis":
		sess.RedisPool = myRedisCache.Conn
	case "mysql", "postgres", "mariadb", "postgresql":
		sess.DBPool = f.DB.Pool
	}

	f.Session = sess.InitSession()
	f.EncryptionKey = os.Getenv("KEY")

	if f.Debug {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
			jet.InDevelopmentMode(),
		)

		f.JetViews = views
	} else {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		)

		f.JetViews = views
	}

	f.createRenderer()
	go f.Mail.ListenForMail()

	return nil
}

func (f *FinishLine) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		// create folder if it doesn't exist
		err := f.CreateDirIfNotExists(root + "/" + path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FinishLine) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		ErrorLog:     f.ErrorLog,
		Handler:      f.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if f.DB.Pool != nil {
		defer f.DB.Pool.Close()
	}

	if redisPool != nil {
		defer redisPool.Close()
	}

	if badgerConn != nil {
		defer badgerConn.Close()
	}

	f.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	err := srv.ListenAndServe()
	f.ErrorLog.Fatal(err)
}

func (f *FinishLine) checkDotEnv(path string) error {
	err := f.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path))

	if err != nil {
		return err
	}
	return nil
}

func (f *FinishLine) startLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}

func (f *FinishLine) createRenderer() {
	myRenderer := render.Render{
		Renderer: f.config.renderer,
		RootPath: f.RootPath,
		Port:     f.config.port,
		JetViews: f.JetViews,
		Session:  f.Session,
	}
	f.Render = &myRenderer
}

func (f *FinishLine) createMailer() mailer.Mail {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	m := mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Templates:   f.RootPath + "/mail",
		Host:        os.Getenv("SMTP_HOST"),
		Port:        port,
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		Encryption:  os.Getenv("SMTP_ENCRYPTION"),
		FromName:    os.Getenv("FROM_NAME"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
		Jobs:        make(chan mailer.Message, 20),
		Results:     make(chan mailer.Result, 20),
		API:         os.Getenv("MAILER_API"),
		APIKey:      os.Getenv("MAILER_KEY"),
		APIUrl:      os.Getenv("MAILER_URL"),
	}

	return m
}

func (f *FinishLine) createClientRedisCache() *cache.RedisCache {
	cacheClient := cache.RedisCache{
		Conn:   f.createRedisPool(),
		Prefix: f.config.redis.prefix,
	}
	return &cacheClient
}

func (f *FinishLine) createClientBadgerCache() *cache.BadgerCache {
	cacheClient := cache.BadgerCache{
		Conn: f.createBadgerConn(),
	}
	return &cacheClient
}

func (f *FinishLine) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				f.config.redis.host,
				redis.DialPassword(f.config.redis.password))
		},

		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}
}

func (f *FinishLine) createBadgerConn() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(f.RootPath + "/temp/badger"))
	if err != nil {
		return nil
	}
	return db
}

func (f *FinishLine) BuildDSN() string {
	var dsn string

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"))

		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}

	default:

	}

	return dsn
}
