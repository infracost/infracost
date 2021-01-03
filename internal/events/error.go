package events

type Error struct {
	Label string
	err   error
}

func (e *Error) Error() string {
	return e.err.Error()
}

func NewError(err error, label string) *Error {
	return &Error{
		Label: label,
		err:   err,
	}
}
