package requestx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"gobkd/internal/apperr"
	"gobkd/internal/response"
)

func BindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		handleBindError(c, err)
		return false
	}
	return true
}

func BindQuery(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindQuery(dst); err != nil {
		handleBindError(c, err)
		return false
	}
	return true
}

func BindURI(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindUri(dst); err != nil {
		handleBindError(c, err)
		return false
	}
	return true
}

func handleBindError(c *gin.Context, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		response.FromError(c, apperr.RequestTooLarge(maxBytesErr.Limit))
		return
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		details := make([]response.ErrorField, 0, len(validationErrs))
		for _, item := range validationErrs {
			details = append(details, response.ErrorField{
				Field: item.Field(),
				Rule:  item.Tag(),
			})
		}
		response.FromError(c, apperr.ValidationFailed(details))
		return
	}

	response.FromError(c, apperr.InvalidRequest(err.Error()))
}
