package models

type Network struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	Password string `db:"password"`
}
