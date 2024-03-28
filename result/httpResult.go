package result

import (
	"bytes"
	"encoding/json"
	"gitee.com/i-Things/share/ctxs"
	"gitee.com/i-Things/share/errors"
	"gitee.com/i-Things/share/utils"
	"io/ioutil"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func HttpErr(w http.ResponseWriter, r *http.Request, httpCode int, err error) {
	var code int
	var msg string
	//错误返回
	er := errors.Fmt(err)

	msg = er.GetI18nMsg(ctxs.GetUserCtxNoNil(r.Context()).AcceptLanguage)

	logx.WithContext(r.Context()).Errorf("【http handle err】router:%v err: %v ",
		r.URL.Path, msg)
	httpx.WriteJson(w, httpCode, Error(er.Code, msg))
	code = int(er.Code)
	//将接口的应答结果写入r.Response，为操作日志记录接口提供应答信息
	var temp http.Response
	temp.StatusCode = code
	temp.Status = msg
	r.Response = &temp
}

// Http http返回
func Http(w http.ResponseWriter, r *http.Request, resp any, err error) {
	var code int
	var msg string
	if err == nil {
		//成功返回
		re := Success(resp)
		httpx.WriteJson(w, http.StatusOK, re)
		code = 200
		msg = "success"

	} else {
		//错误返回
		er := errors.Fmt(err)

		msg = er.GetI18nMsg(ctxs.GetUserCtxNoNil(r.Context()).AcceptLanguage)

		logx.WithContext(r.Context()).Errorf("【http handle err】router:%v err: %v ",
			r.URL.Path, msg)
		httpx.WriteJson(w, http.StatusOK, Error(er.Code, msg))
		code = int(er.Code)
	}

	//将接口的应答结果写入r.Response，为操作日志记录接口提供应答信息
	bs, _ := json.Marshal(resp)
	var temp http.Response
	temp.StatusCode = code
	temp.Status = msg
	temp.Body = ioutil.NopCloser(bytes.NewReader(bs))
	r.Response = &temp
}

// HttpWithoutWrap http返回，无包装
func HttpWithoutWrap(w http.ResponseWriter, r *http.Request, resp any, err error) {
	if err == nil {
		//成功返回
		httpx.WriteJson(w, http.StatusOK, resp)
	} else {
		//错误返回
		er := errors.Fmt(err)
		logx.WithContext(r.Context()).Errorf("【http handle err】router:%v err: %v ",
			r.URL.Path, utils.Fmt(er))
		httpx.WriteJson(w, http.StatusOK, Error(er.Code, er.GetMsg()))
	}
}
