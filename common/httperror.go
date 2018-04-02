package common

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/mtdx/ns-ga/logger"
)

// ErrResponse ...
type ErrResponse struct {
	Err            error `json:"-"`    // low-level runtime error
	HTTPStatusCode int   `json:"code"` // http response status code

	StatusText string `json:"status"`            // user-level status message
	AppCode    int64  `json:"appcode,omitempty"` // application-specific error code
	ErrorText  string `json:"error,omitempty"`   // application-level error message, for debugging
}

// Render ...
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrInvalidRequest ...
func ErrInvalidRequest(err error) render.Renderer {
	logger.Info.Println(err.Error())
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     err.Error(),
	}
}

// ErrInternalServer ...
func ErrInternalServer(err error) render.Renderer {
	logger.Error.Println(err.Error())
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 500,
		StatusText:     err.Error(),
	}
}

// ErrRender ...
func ErrRender(err error) render.Renderer {
	logger.Warning.Println(err.Error())
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     err.Error(),
	}
}

// ErrNotFound ...
var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource Not Found"}
