package db

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
func (m *Metadata) Query() error {
	return Db.Where("name = ?", m.Name).First(&m).Error
}
