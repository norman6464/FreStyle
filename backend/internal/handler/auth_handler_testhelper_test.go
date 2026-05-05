package handler

import (
	"net/http"
	"net/http/httptest"
)

// mustNewRequest はテスト用に最小の *http.Request を返す。
// auth_handler 配下の handler は c.Request.Context() しか参照しないため、URL / method は何でも良い。
func mustNewRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/", nil)
}
