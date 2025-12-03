package http

import (
	"net/http"

	"github.com/krateoplatformops/plumbing/http/response"
)

// Status is a return value for calls that don't return other objects.
type Status struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`

	// Status of the operation.
	// One of: "Success" or "Failure".
	Status string `json:"status,omitempty"`
	// A human-readable description of the status of this operation.
	Message string `json:"message,omitempty"`
	// A machine-readable description of why this operation is in the
	// "Failure" status. If this value is empty there
	// is no information available. A Reason clarifies an HTTP status
	// code but does not override it.
	Reason response.StatusReason `json:"reason,omitempty"`
	// Suggested HTTP return code for this status, 0 if not set.
	Code int `json:"code,omitempty"`
	// Response Headers
	Header *http.Header `json:"header,omitempty"`
}

func New(code int, header *http.Header, err error) *Status {
	res := &Status{
		Kind:       "Status",
		APIVersion: "v1",
		Code:       code,
	}

	if err != nil {
		res.Message = err.Error()
	}

	switch code {
	case http.StatusUnprocessableEntity:
		res.Status = response.StatusFailure
		res.Reason = response.StatusUnprocessableEntity
	case http.StatusUnauthorized:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonUnauthorized
	case http.StatusForbidden:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonForbidden
	case http.StatusNotFound:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonNotFound
	case http.StatusConflict:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonConflict
	case http.StatusGone:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonGone
	case http.StatusNotImplemented:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonInvalid
	case http.StatusBadRequest:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonBadRequest
	case http.StatusServiceUnavailable:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonServiceUnavailable
	case http.StatusNotAcceptable:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonNotAcceptable
	case http.StatusMethodNotAllowed:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonMethodNotAllowed
	case http.StatusInternalServerError:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonInternalError
	case http.StatusRequestEntityTooLarge:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonRequestEntityTooLarge
	case http.StatusUnsupportedMediaType:
		res.Status = response.StatusFailure
		res.Reason = response.StatusReasonUnsupportedMediaType
	default:
		res.Status = response.StatusSuccess
		res.Header = header
	}

	return res
}
