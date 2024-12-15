package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

const (
	StatusOk  = "Ok"
	StatusErr = "Error"
)

type Response struct {
	Status  string           `json:"status"`
	Message string           `json:"message,omitempty"`
	Error   string           `json:"error,omitempty"`
	Errors  validationErrors `json:"errors,omitempty"`
}

type validationErrors map[string]string

func RespOk(msg string) *Response {
	return &Response{
		Status:  StatusOk,
		Message: msg,
	}
}

func RespErr(msg string) *Response {
	return &Response{
		Status:  StatusErr,
		Message: msg,
	}
}

func ErrInternal(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusInternalServerError)
	render.Render(w, r, Response{ //nolint:errcheck
		Status:  StatusErr,
		Message: "Internal error",
	})
}

func ErrUnprocessableEntity(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusUnprocessableEntity)
	render.Render(w, r, Response{ //nolint:errcheck
		Status:  StatusErr,
		Message: "Unprocessable entity",
	})
}

func ErrConflict(w http.ResponseWriter, r *http.Request, msg string) {
	render.Status(r, http.StatusConflict)
	render.Render(w, r, Response{ //nolint:errcheck
		Status:  StatusErr,
		Message: msg,
	})
}

func ErrBadRequest(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusBadRequest)
	render.Render(w, r, Response{ //nolint:errcheck
		Status:  StatusErr,
		Message: "Bad request",
	})
}

func ErrUnauthorized(w http.ResponseWriter, r *http.Request, msg string) {
	render.Status(r, http.StatusUnauthorized)
	render.Render(w, r, Response{ //nolint:errcheck
		Status:  StatusErr,
		Message: msg,
	})
}

func ErrInvalid(w http.ResponseWriter, r *http.Request, err error) {
	errs := make(validationErrors)

	for _, e := range err.(validator.ValidationErrors) {
		errs[strings.ToLower(e.Field())] = fmt.Sprintf("field must satisfy '%s' constraint", e.Tag())
	}

	render.Status(r, http.StatusBadRequest)
	render.Render(w, r, Response{ //nolint:errcheck
		Status:  StatusErr,
		Message: "Some fields are invalid",
		Errors:  errs,
	})
}

func (resp Response) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (ve validationErrors) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
