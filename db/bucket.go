package db

import (
	"github.com/jinzhu/gorm"
)

type Bucket struct {
	Model
	AccessKey string
	SecretKey string
	Name      string
	UserID    int
	Host      string
}

func (Bucket) TableName() string {
	return "bucket"
}

//Query .
func (b *Bucket) Query() error {
	err := Db.Model(&b).Where("host = ?", b.Host).First(&b).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}
