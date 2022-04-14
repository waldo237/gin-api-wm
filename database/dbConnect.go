package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/waldo237/gin-api-wm/config"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// Import MySQL database driver
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	"gorm.io/driver/mysql"

	// Import PostgreSQL database driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/driver/postgres"

	// Import SQLite3 database driver
	// _ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/driver/sqlite"

	// Import Redis Driver
	"github.com/mediocregopher/radix/v4"

	// Import Mongo driver
	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/event"
	opts "go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

// DB global variable to access gorm
var DB *gorm.DB

var sqlDB *sql.DB
var err error

// RedisClient global variable to access the redis client
var RedisClient radix.Client

// RedisConnTTL - context deadline in second
var RedisConnTTL int

// MongoClient instance
var MongoClient *qmgo.Client

// InitDB - function to initialize db
func InitDB() *gorm.DB {
	var db = DB

	configureDB := config.Database().RDBMS

	driver := configureDB.Env.Driver
	username := configureDB.Access.User
	password := configureDB.Access.Pass
	database := configureDB.Access.DbName
	host := configureDB.Env.Host
	port := configureDB.Env.Port
	sslmode := configureDB.Ssl.Sslmode
	timeZone := configureDB.Env.TimeZone
	maxIdleConns := configureDB.Conn.MaxIdleConns
	maxOpenConns := configureDB.Conn.MaxOpenConns
	connMaxLifetime := configureDB.Conn.ConnMaxLifetime
	logLevel := configureDB.Log.LogLevel

	switch driver {
	case "mysql":
		dsn := username + ":" + password + "@tcp(" + host + ":" + port + ")/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"
		sqlDB, err = sql.Open(driver, dsn)
		if err != nil {
			log.WithError(err).Panic("panic code: 151")
		}
		sqlDB.SetMaxIdleConns(maxIdleConns)       // max number of connections in the idle connection pool
		sqlDB.SetMaxOpenConns(maxOpenConns)       // max number of open connections in the database
		sqlDB.SetConnMaxLifetime(connMaxLifetime) // max amount of time a connection may be reused

		db, err = gorm.Open(mysql.New(mysql.Config{
			Conn: sqlDB,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(logLevel)),
		})
		if err != nil {
			log.WithError(err).Panic("panic code: 152")
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	case "postgres":
		dsn := "host=" + host + " port=" + port + " user=" + username + " dbname=" + database + " password=" + password + " sslmode=" + sslmode + " TimeZone=" + timeZone
		sqlDB, err = sql.Open(driver, dsn)
		if err != nil {
			log.WithError(err).Panic("panic code: 153")
		}
		sqlDB.SetMaxIdleConns(maxIdleConns)       // max number of connections in the idle connection pool
		sqlDB.SetMaxOpenConns(maxOpenConns)       // max number of open connections in the database
		sqlDB.SetConnMaxLifetime(connMaxLifetime) // max amount of time a connection may be reused

		db, err = gorm.Open(postgres.New(postgres.Config{
			Conn: sqlDB,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(logLevel)),
		})
		if err != nil {
			log.WithError(err).Panic("panic code: 154")
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(database), &gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Silent),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			log.WithError(err).Panic("panic code: 155")
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	default:
		log.Fatal("The driver " + driver + " is not implemented yet")
	}

	DB = db

	return DB
}

// GetDB - get a connection
func GetDB() *gorm.DB {
	return DB
}

// InitRedis - function to initialize redis client
func InitRedis() (radix.Client, error) {
	configureRedis := config.Database().REDIS
	RedisConnTTL = configureRedis.Conn.ConnTTL

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(RedisConnTTL)*time.Second)
	defer cancel()

	rClient, err := (radix.PoolConfig{
		Size: configureRedis.Conn.PoolSize,
	}).New(ctx, "tcp", fmt.Sprintf("%v:%v",
		configureRedis.Env.Host,
		configureRedis.Env.Port))
	if err != nil {
		log.WithError(err).Panic("panic code: 161")
		return rClient, err
	}
	// Only for debugging
	if err == nil {
		fmt.Println("REDIS pool connection successful!")
	}

	RedisClient = rClient

	return RedisClient, nil
}

// GetRedis - get a connection
func GetRedis() radix.Client {
	return RedisClient
}

// InitMongo - function to initialize mongo client
func InitMongo() (*qmgo.Client, error) {
	configureMongo := config.Database().MongoDB

	// Connect to the database or cluster
	URI := configureMongo.Env.URI

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configureMongo.Env.ConnTTL)*time.Second)
	defer cancel()

	clientConfig := &qmgo.Config{
		Uri:         URI,
		MaxPoolSize: &configureMongo.Env.PoolSize,
	}
	serverAPIOptions := opts.ServerAPI(opts.ServerAPIVersion1)

	opt := opts.Client().SetAppName(configureMongo.Env.AppName)
	opt.SetServerAPIOptions(serverAPIOptions)

	// for monitoring pool
	if configureMongo.Env.PoolMon == "yes" {
		poolMonitor := &event.PoolMonitor{
			Event: func(evt *event.PoolEvent) {
				switch evt.Type {
				case event.GetSucceeded:
					fmt.Println("GetSucceeded")
				case event.ConnectionReturned:
					fmt.Println("ConnectionReturned")
				}
			},
		}
		opt.SetPoolMonitor(poolMonitor)
	}

	client, err := qmgo.NewClient(ctx, clientConfig, options.ClientOptions{ClientOptions: opt})
	if err != nil {
		return client, err
	}

	// Only for debugging
	fmt.Println("MongoDB pool connection successful!")

	MongoClient = client

	return MongoClient, nil
}

// GetMongo - get a connection
func GetMongo() *qmgo.Client {
	return MongoClient
}
