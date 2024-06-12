package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

// implements the MarshalJSON() method on Runtime so that
// it satisfies the json.Marshaler interface
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	quotedValue := strconv.Quote(jsonValue)
	return []byte(quotedValue), nil
}

// implements the UnmarshalJSON() method on Runtime so that
// it satisfies the json.Unmarshler interface
// Note: uses *Runtime instead of Runtime because it modifies the receiver
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// the incoming JSON value will be a string of the format "<runtime> mins"
	unquotesJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	parts := strings.Split(unquotesJSONValue, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	value, err := strconv.Atoi(parts[0])
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(value)
	return nil
}
