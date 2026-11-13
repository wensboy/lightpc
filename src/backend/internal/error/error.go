package error

import "fmt"

const (
	ErrorRoot = '1' // Code[0]
)

// Terror 是一个可追踪的 error 结构, 实现了标准 Error 和 Unwrap() error 接口.
type Terror struct {
	Code   string
	Kind   string
	Msg    string
	source error
}

type Option struct {
	Code string
	Kind string
	Msg  string
}

func New(opt Option) *Terror {
	return &Terror{
		Code: opt.Code,
		Kind: opt.Kind,
		Msg:  opt.Msg,
	}
}

func (e *Terror) Wrap(err error) *Terror {
	e.source = err
	return e
}

func (e *Terror) Error() string {
	out := fmt.Sprintf("[%s|%s] %s", e.Code, e.Kind, e.Msg)
	if e.Code[0] == ErrorRoot {
		return out
	}
	return fmt.Sprintf("%s -> %s", out, e.source.Error())
}

func (e *Terror) Unwrap() error {
	return e.source
}
