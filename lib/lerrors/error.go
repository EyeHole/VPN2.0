package lerrors

type Error struct {
	Err error
	NetID int
	ClientID int
}

func (e Error) Error() string {
	return e.Err.Error()
}