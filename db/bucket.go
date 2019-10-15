package db

type Bucket struct {
	Model
	AccessKey string
	SecretKey string
	Name      string
	UserID    int
}

func (Bucket) TableName() string {
	return "bucket"
}
