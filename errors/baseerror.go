package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"runtime"
	"runtime/debug"
)

type CodeError struct {
	Code    int64      `json:"code"`
	Msg     []I18nImpl `json:"-"`
	MsgStr  string     `json:"msg"`
	Details []string   `json:"details,omitempty"`
	Stack   []string   `json:"stack,omitempty"`
}

type RpcError interface {
	GRPCStatus() *status.Status
}

//func TogRPCError(err *Error) error {
//	s, _ := status.New(ToRPCCode(err.Code()), err.Msg()).WithDetails(&pb.Error{Code: int32(err.Code()), Message: err.Msg()})
//	return s.Err()
//}

func (c CodeError) ToRpc(accept string) error {
	code := codes.Unknown
	switch c.Code {
	case Failure.Code: //失败需要回滚
		code = codes.Aborted
	case OnGoing.Code: //任务还在执行中
		code = codes.FailedPrecondition
	}
	c.MsgStr = c.GetI18nMsg(accept)
	s := status.New(code, c.Error())
	return s.Err()
}

func ToRpc(err error, accept string) error {
	if err == nil {
		return err
	}
	switch err.(type) {
	case RpcError:
		return err
	case *CodeError:
		return err.(*CodeError).ToRpc(accept)
	default:
		return Fmt(err).ToRpc(accept)
	}
}

func (c CodeError) WithMsg(msg string) *CodeError {

	c.Msg = []I18nImpl{String(msg)}
	return &c
}

func (c CodeError) WithMsgf(format string, a ...any) *CodeError {
	c.Msg = []I18nImpl{Msgf{Format: format, A: a}}
	return &c
}

func (c CodeError) AddMsg(msg string) *CodeError {
	c.Msg = append(c.Msg, String(msg))
	return &c
}

func (c CodeError) AddMsgf(format string, a ...any) *CodeError {
	c.Msg = append(c.Msg, Msgf{Format: format, A: a})
	return &c
}

func (c CodeError) AddDetail(msg ...any) *CodeError {
	c.Details = append(c.Details, fmt.Sprint(msg...))
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	c.Stack = append(c.Stack, string(debug.Stack()))
	return &c
}
func (c CodeError) WithStack() *CodeError {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	c.Stack = append(c.Stack, string(debug.Stack()))
	return &c
}

func (c CodeError) AddDetailf(format string, a ...any) *CodeError {
	c.Details = append(c.Details, fmt.Sprintf(format, a...))
	return &c
}

func (c *CodeError) GetDetailMsg() string {
	if c == nil {
		return OK.GetDetailMsg()
	}
	if len(c.Details) == 0 {
		return c.GetMsg()
	}
	return fmt.Sprintf("msg=%s,detail=%v", c.Msg, c.Details)
}

func (c *CodeError) GetCode() int64 {
	if c == nil { //如果没错误,则是成功
		return OK.Code
	}
	return c.Code
}

func (c *CodeError) GetMsg() string {
	if c == nil {
		return OK.GetMsg()
	}
	return stringMsgs(c.Msg)
}

var ErrorMap = map[int64]string{}

func NewCodeError(code int64, msg string) *CodeError {
	ErrorMap[code] = msg
	return &CodeError{Code: code, Msg: []I18nImpl{String(msg)}}
}

func NewDefaultError(msg string) error {
	return Default.WithMsg(msg)
}

func (c CodeError) Error() string {
	c.Stack = nil
	c.MsgStr = c.GetI18nMsg("")
	ret, _ := json.Marshal(c)
	return string(ret)
}

// 将普通的error及转换成json的error或error类型的转回自己的error
func Fmt(errs error) *CodeError {
	if errs == nil {
		return nil
	}
	switch errs.(type) {
	case *CodeError:
		return errs.(*CodeError)
	case RpcError: //如果是grpc类型的错误
		s, _ := status.FromError(errs)
		if s.Code() != codes.Unknown { //只有自定义的错误,grpc会返回unknown错误码
			err := fmt.Sprintf("rpc err detail is nil|err=%#v", s)
			return System.AddDetail(err)
		}
		var ret CodeError
		err := json.Unmarshal([]byte(s.Message()), &ret)
		ret.Msg = []I18nImpl{String(ret.MsgStr)}
		if err != nil {
			return System.AddDetail(err)
		}
		return &ret
	case *goja.Exception:
		e := errs.(*goja.Exception)
		return Script.AddMsg(e.Error())
	default:
		var ce CodeError
		err := json.Unmarshal([]byte(errs.Error()), &ce)
		if err != nil {
			return System.AddDetail(errs.Error())
		}
		return Default.AddDetail(errs)
	}
}

func Cmp(err1 error, err2 error) bool {
	if err2 == nil && err1 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return Fmt(err1).Code == Fmt(err2).Code
}
func IfNotNil(c *CodeError, err error) error {
	if err != nil {
		return c.AddDetail(err)
	}
	return nil
}
func Is(err, target error) bool {
	return errors.Is(err, target)
}

func Must(err error, msg string) {
	if err != nil {
		pc := make([]uintptr, 1)
		runtime.Callers(2, pc)
		stack := string(debug.Stack())
		logx.Errorf("出现一个程序退出错误:%v,err:%v,stack:%v", msg, err, stack)
		os.Exit(-1)
	}
}
