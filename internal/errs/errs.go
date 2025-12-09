package errs

type Err string

func (e Err) Error() string {
	return string(e)
}

const (
	InvalidToken = Err("invalid_token")
	ServiceNA    = Err("service_not_available")
)

// ErrFull

type ErrFull struct {
	Err    error
	Desc   string
	Fields map[string]string
}

func (e ErrFull) Error() string {
	return e.Err.Error() + ", desc:" + e.Desc
}
