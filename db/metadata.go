package db

import (
	"github.com/jinzhu/gorm"
)

//Metadata 元数据.
type Metadata struct {
	Model
	Name      string `gorm:"index"`
	Type      string
	Size      int64
	Hash      string `gorm:"index"`
	StoreNode string
	Usable    bool `gorm:"index"` //节点是否可用.
}

//TableName .
func (Metadata) TableName() string {
	return "metadata"
}

//Create 创建.
func (m *Metadata) Create() error {
	return Db.Create(&m).Error
}

//Query.
func (m *Metadata) QueryNodeIP() (list []string, err error) {
	metaList := []Metadata{}
	err = Db.Where("name = ?", m.Name).Find(&metaList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return list, err
	}

	for _, v := range metaList {
		list = append(list, v.StoreNode)
		m.Size = v.Size
		m.Hash = v.Hash
	}

	return list, nil
}
