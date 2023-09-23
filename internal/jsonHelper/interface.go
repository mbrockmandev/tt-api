package jsonHelper

import (
	"net/http"
)

type IJsonHelper interface {
	WriteJson(w http.ResponseWriter, status int,
		data interface{},
		headers ...http.Header) error
	ReadJson(w http.ResponseWriter, r *http.Request,
		data interface{}) error
	ErrorJson(w http.ResponseWriter, err error,
		status ...int) error
}
