package models

type Network struct {
	ID       int    `db:"id"`
	NetName  string `db:"net_name"`
	Password string `db:"password"`
	Mask     int    `db:"mask"`
}
