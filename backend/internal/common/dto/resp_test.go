package dto

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccessNoData(t *testing.T) {
	resp := SuccessNoData()
	assert.Equal(t, uint(http.StatusOK), resp.Code)
	assert.Equal(t, "success", resp.Msg)
	assert.Nil(t, resp.Data)
}

func TestSuccess(t *testing.T) {
	data := "test-data"
	resp := Success(data)
	assert.Equal(t, uint(http.StatusOK), resp.Code)
	assert.Equal(t, "success", resp.Msg)
	assert.Equal(t, data, resp.Data)
}

func TestSuccessWithStruct(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}
	user := User{Name: "Alice", Age: 30}
	resp := Success(user)
	assert.Equal(t, uint(http.StatusOK), resp.Code)
	assert.Equal(t, "success", resp.Msg)
	assert.Equal(t, user, resp.Data)
}

func TestSuccessWithSlice(t *testing.T) {
	items := []string{"a", "b", "c"}
	resp := Success(items)
	assert.Equal(t, uint(http.StatusOK), resp.Code)
	assert.Equal(t, items, resp.Data)
}

func TestFailBadReq(t *testing.T) {
	msg := "bad request"
	resp := FailBadReq(msg)
	assert.Equal(t, uint(http.StatusBadRequest), resp.Code)
	assert.Equal(t, msg, resp.Msg)
	assert.Nil(t, resp.Data)
}

func TestFailInternalServerError(t *testing.T) {
	msg := "internal error"
	resp := FailInternalServerError(msg)
	assert.Equal(t, uint(http.StatusInternalServerError), resp.Code)
	assert.Equal(t, msg, resp.Msg)
	assert.Nil(t, resp.Data)
}

func TestFailForbidden(t *testing.T) {
	msg := "forbidden"
	resp := FailForbidden(msg)
	assert.Equal(t, uint(http.StatusForbidden), resp.Code)
	assert.Equal(t, msg, resp.Msg)
	assert.Nil(t, resp.Data)
}

func TestFailUnauthorized(t *testing.T) {
	msg := "unauthorized"
	resp := FailUnauthorized(msg)
	assert.Equal(t, uint(http.StatusUnauthorized), resp.Code)
	assert.Equal(t, msg, resp.Msg)
	assert.Nil(t, resp.Data)
}

func TestFailNotFound(t *testing.T) {
	msg := "not found"
	resp := FailNotFound(msg)
	assert.Equal(t, uint(http.StatusNotFound), resp.Code)
	assert.Equal(t, msg, resp.Msg)
	assert.Nil(t, resp.Data)
}

func TestFailTooManyReq(t *testing.T) {
	msg := "too many requests"
	resp := FailTooManyReq(msg)
	assert.Equal(t, uint(http.StatusTooManyRequests), resp.Code)
	assert.Equal(t, msg, resp.Msg)
	assert.Nil(t, resp.Data)
}

func TestFail(t *testing.T) {
	code := uint(400)
	msg := "custom error"
	resp := Fail(code, msg)
	assert.Equal(t, code, resp.Code)
	assert.Equal(t, msg, resp.Msg)
	assert.Nil(t, resp.Data)
}

func TestAPIResp_TypeParameters(t *testing.T) {
	respInt := Success(42)
	assert.Equal(t, 42, respInt.Data)

	respStr := Success("hello")
	assert.Equal(t, "hello", respStr.Data)

	// Test generic APIResp with different types
	intAPI := APIResp[int]{Code: 200, Msg: "ok", Data: 100}
	assert.Equal(t, 100, intAPI.Data)
}
