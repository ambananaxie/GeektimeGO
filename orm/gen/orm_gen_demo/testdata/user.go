package testdata

import sqlx "database/sql"

type User struct {
	Name     string
	Age      *int
	NickName *sqlx.NullString
	Picture  []byte
}

type UserDetail struct {
	Address string
}
