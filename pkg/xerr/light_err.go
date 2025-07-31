package xerr

import "fmt"

type LightErr struct {
	cause error
	Msg   string
}

func (e LightErr) Error() string {
	return fmt.Sprintf("%s - cause: %v", e.Msg, e.cause)
}

func Wrap(err error, msg string, args ...any) error {
	if err == nil {
		return nil
	}
	return &LightErr{
		cause: err,
		Msg:   fmt.Sprintf(msg, args...),
	}
}
