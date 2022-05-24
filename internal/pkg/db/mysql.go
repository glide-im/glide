package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

var DB *gorm.DB

type MySQLConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	Database       string
	Charset        string
	MaxOpenConn    int
	MaxIdleConn    int
	MaxLifeTimeMin int
}

func Init(mysqlConf *MySQLConfig, redisConf *RedisConfig) error {

	if mysqlConf != nil {
		err := initMySQL(mysqlConf)
		if err != nil {
			return err
		}
	}
	if redisConf != nil {
		initRedis(redisConf)
	}
	return nil
}

func initMySQL(mysqlConf *MySQLConfig) error {

	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true",
		mysqlConf.User, mysqlConf.Password, mysqlConf.Host, mysqlConf.Port, mysqlConf.Database, mysqlConf.Charset)

	var err error
	DB, err = gorm.Open(mysql.Open(url), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "im_",
			SingularTable: true,
			//NameReplacer:  nil,
			//NoLowerCase:   false,
		},
	})
	if err != nil {
		return err
	}
	db, err := DB.DB()
	if err != nil {
		return err
	}
	if mysqlConf.MaxOpenConn > 0 {
		db.SetMaxOpenConns(mysqlConf.MaxOpenConn)
	}
	if mysqlConf.MaxLifeTimeMin > 0 {
		db.SetConnMaxLifetime(time.Duration(mysqlConf.MaxLifeTimeMin) * time.Minute)
	}
	if mysqlConf.MaxIdleConn > 0 {
		db.SetMaxIdleConns(mysqlConf.MaxIdleConn)
	}
	return nil
}
