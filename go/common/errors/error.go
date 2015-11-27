package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

var (
	ERR_NO_DATA = New("have no data info", "errors.ERR_NO_DATA")
)

func test() error {
	return errors.New("test")
}

type ErrorT struct {
	Code   string   `json:"code"`
	Caller []string `json:"caller"`
	Reason []string `json:"reason"`
}

func caller(dept int) string {
	_, file, line, ok := runtime.Caller(dept)
	if !ok {
		return "Unknown"
	}

	idx := strings.LastIndex(file, "/")

	return fmt.Sprint(file[idx+1:], ":", line)

}

func Code(err error) int {
	code := err.Error()
	if e, ok := err.(*ErrorT); ok {
		code = e.Code
	}

	codeInt, err := strconv.ParseInt(code, 10, 32)
	if err != nil {
		println("error code parse error:", err)
		return 1
	}

	return int(codeInt)
}

//追加错误
func As(err error, reason ...interface{}) error {
	if err == nil {
		return nil
	}

	e, ok := err.(*ErrorT)
	if !ok {
		return &ErrorT{
			Code:   err.Error(),
			Caller: []string{caller(2)},
			Reason: []string{fmt.Sprint(reason)},
		}
	}
	e.Caller = append(e.Caller, caller(2))
	e.Reason = append(e.Reason, fmt.Sprint(reason))
	return e
}

func Equal(e1 error, e2 error) bool {
	if e1 == e2 {
		return true
	}
	err1, ok1 := e1.(*ErrorT)
	err2, ok2 := e2.(*ErrorT)
	if ok1 && ok2 && err1.Equal(err2) {
		return true
	}

	return e1.Error() == e2.Error()
}

// 新建错误
func New(code string, reason ...interface{}) *ErrorT {
	return &ErrorT{
		Code:   code,
		Caller: []string{caller(2)},
		Reason: []string{fmt.Sprint(reason)},
	}
}

// 追加错误
func (e *ErrorT) As(reason ...interface{}) *ErrorT {
	e.Caller = append(e.Caller, caller(2))
	e.Reason = append(e.Reason, fmt.Sprint(reason))
	return e
}

// 复制
func (e *ErrorT) Clone() *ErrorT {
	return &ErrorT{
		Code:   e.Code,
		Caller: append([]string{}, e.Caller...),
		Reason: append([]string{}, e.Reason...),
	}
}

// error 接口实现
func (e *ErrorT) Error() string {
	count := len(e.Caller)
	errStr := fmt.Sprint("\nerr_code:", e.Code)
	for i := 0; i < count; i++ {
		errStr += fmt.Sprintf("\nerr_stack_%v==> caller:%v reason:%v", i, e.Caller[i], e.Reason[i])
	}
	return errStr
}

// err 是否相同
func (e *ErrorT) Equal(err error) bool {
	if e1, ok := err.(*ErrorT); ok {
		return e.Code == e1.Code
	}
	return e.Error() == err.Error()
}
