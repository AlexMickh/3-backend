package response

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

type HttpError struct {
	msg    string
	status int
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func ErrorWrapper(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			w.Header().Add("Content-Type", "application/json")
			httpErr, ok := err.(HttpError)
			if ok {
				w.WriteHeader(httpErr.status)
				render.JSON(w, r, ErrorResponse{Error: httpErr.msg})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, ErrorResponse{Error: err.Error()})
			}
		}
	}
}

func Error(msg string, status int) HttpError {
	return HttpError{
		msg:    msg,
		status: status,
	}
}

func (h HttpError) Error() string {
	return fmt.Sprintf("status: %d, msg: %s", h.status, h.msg)
}
