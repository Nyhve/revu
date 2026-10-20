package httpx

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxBodyBytes = 1 << 20

// Validatable request bodies check themselves after decoding.
type Validatable interface {
	Validate() error
}

// Bind decodes the JSON body into dst and validates it.
// Unknown fields are ignored (as Spring/Jackson did).
func Bind(c *gin.Context, dst any) error {
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes))
	if err != nil {
		return BadRequest("Request body could not be read")
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return BadRequest("Required request body is missing")
	}
	if err := json.Unmarshal(body, dst); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) && typeErr.Field != "" {
			return BadRequest("Malformed JSON request: invalid value for field '%s'", typeErr.Field)
		}
		return BadRequest("Malformed JSON request")
	}
	if v, ok := dst.(Validatable); ok {
		return v.Validate()
	}
	return nil
}
