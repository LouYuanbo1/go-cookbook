package dto

import "net/http"

type APIResp[T any] struct {
	Code uint   `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

func SuccessNoData() APIResp[any] {
	return Success[any](nil)
}

func Success[T any](data T) APIResp[T] {
	return APIResp[T]{Code: http.StatusOK, Msg: "success", Data: data}
}

func FailBadReq(msg string) APIResp[any] {
	return Fail(http.StatusBadRequest, msg)
}

func FailInternalServerError(msg string) APIResp[any] {
	return Fail(http.StatusInternalServerError, msg)
}

func FailForbidden(msg string) APIResp[any] {
	return Fail(http.StatusForbidden, msg)
}

func FailUnauthorized(msg string) APIResp[any] {
	return Fail(http.StatusUnauthorized, msg)
}

func FailNotFound(msg string) APIResp[any] {
	return Fail(http.StatusNotFound, msg)
}

func FailTooManyReq(msg string) APIResp[any] {
	return Fail(http.StatusTooManyRequests, msg)
}

func Fail(code uint, msg string) APIResp[any] {
	return APIResp[any]{Code: code, Msg: msg, Data: nil}
}
