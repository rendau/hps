package errs

type Err string

func (e Err) Error() string {
	return string(e)
}

const (
	ServerNA = Err("server_not_available")
)
