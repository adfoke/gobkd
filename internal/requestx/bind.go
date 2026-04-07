package requestx

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

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
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		details := make([]response.ErrorField, 0, len(validationErrs))
		for _, item := range validationErrs {
			details = append(details, response.ErrorField{
				Field: item.Field(),
				Rule:  item.Tag(),
			})
		}
		response.ValidationFailed(c, details)
		return
	}

	response.InvalidRequest(c, err.Error())
}
