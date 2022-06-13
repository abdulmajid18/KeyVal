package data

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/abdulmajid18/keyVal/key_value/internal/validator"
	"github.com/abdulmajid18/keyVal/key_value/other/helper"
)

type PutModel struct {
	DB *sql.DB
}

type PutData struct {
	DbName string `json:"dbname"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

func ValidatePutData(v *validator.Validator, data *PutData) {
	v.Check(data.DbName == "", "Database Name", "must be provided")
	v.Check(len(data.DbName) >= 300, "Datbase Name", "must not be more than 300 bytes long")

	v.Check(data.Key == "", " Key", "must be provided")
	v.Check(len(data.Key) >= 30, " Key", "must not be more than 30 bytes long")

	v.Check(data.Value == "", " Value", "must be provided")
	v.Check(len(data.Value) >= 90, " Value", "must not be more than 90 bytes long")
}

func (m PutModel) CheckExistenceDB(secretKey string) (bool, error) {
	var db string
	query := `SELECT dbname FROM  users WHERE dbname = $1`

	err := m.DB.QueryRow(query, secretKey).Scan(&db)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, ErrRecordNotFound
		default:
			return false, err
		}
	}
	return true, nil
}

func OpenDB(dbname string) (*helper.DB, error) {
	new_db := fmt.Sprintf("%s.db", dbname)

	path := fmt.Sprintf("/home/rozz/Desktop/database/%s", new_db)
	db, err := helper.Open(path)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Insert(data PutData) error {
	db, err := OpenDB(data.DbName)
	if err != nil {
		return err
	}
	err = db.Put(data.Key, data.Value)
	if err != nil {
		return err
	}
	return nil
}
