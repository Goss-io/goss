package db

type User struct {
	Model
	Name     string
	Account  string
	Password string
}

//TableName.
func (User) TableName() string {
	return "user"
}
