package data

import (
	"database/sql"
)

// var (
// 	ErrRecordNotFound = errors.New("record not found")
// )

type Models struct {
	Users UserModel
	Put   PutModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB: db},
		Put:   PutModel{DB: db},
	}
}
