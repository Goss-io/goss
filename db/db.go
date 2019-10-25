package db

import (
	"fmt"
	"log"

	"github.com/Goss-io/goss/lib/ini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

//Db .
var Db *gorm.DB

//DbConfig .
type DbConfig struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     int
	Charset  string
}

//Init .
func Connection() error {
	cf := DbConfig{
		Host:     ini.GetString("db_host"),
		User:     ini.GetString("db_user"),
		Password: ini.GetString("db_pwd"),
		Name:     ini.GetString("db_name"),
		Port:     ini.GetInt("db_port"),
		Charset:  ini.GetString("db_charset"),
	}

	return conndb(cf)
}

//conndb .
func conndb(cf DbConfig) error {
	args := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true",
		cf.User,
		cf.Password,
		cf.Host,
		cf.Port,
		cf.Name,
	)

	log.Println("args:", args)
	db, err := gorm.Open("mysql", args)
	if err != nil {
		log.Printf("%+v\n", err)
		return err
	}

	db.SingularTable(true)
	db.LogMode(true)

	autoMigrate(db)
	Db = db

	return nil
}

//autoMigrate .
func autoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		&User{},
		&Bucket{},
		&Metadata{},
	)
}
