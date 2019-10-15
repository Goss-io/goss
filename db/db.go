package db

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/Goss-io/goss/lib/ini"
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
	args := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cf.Host,
		cf.Port,
		cf.User,
		cf.Password,
		cf.Name,
	)

	log.Println("args:", args)
	db, err := gorm.Open("postgres", args)
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
