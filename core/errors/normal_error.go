package errors

type NormalError struct {
	Message string
}

func (e *NormalError) Error() string {
	return e.Message
}
