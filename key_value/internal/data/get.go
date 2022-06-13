package data

import (
	"github.com/abdulmajid18/keyVal/key_value/internal/validator"
)

type GetData struct {
	DbName string `json:"dbname"`
	Key    string `json:"key"`
}

func ValidateGetData(v *validator.Validator, data *GetData) {
	v.Check(data.DbName == "", "Database Name", "must be provided")
	v.Check(len(data.DbName) >= 300, "Datbase Name", "must not be more than 300 bytes long")

	v.Check(data.Key == "", " Key", "must be provided")
	v.Check(len(data.Key) >= 30, " Key", "must not be more than 30 bytes long")
}

func Get(data GetData) (string, bool, error) {
	db, err := OpenDB(data.DbName)
	if err != nil {
		return "", false, err
	}
	value, state, err := db.Get(data.Key)
	if err != nil {
		return "", false, err
	}
	return value, state, nil
}
