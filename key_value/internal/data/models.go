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
	Token TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB: db},
		Put:   PutModel{DB: db},
		Token: TokenModel{DB: db},
	}
}
