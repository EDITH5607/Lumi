package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32


//When Go is encoding a particular type to JSON, it looks to see if the type has a MarshalJSON()
// method implemented on it. If it has, then Go will call this method to determine how to encode
// it


//type Marshaler interface {
// MarshalJSON() ([]byte, error)
// }
// If the type does satisfy the interface, then Go will call its MarshalJSON() method and use the
// []byte slice that it returns as the encoded JSON value.
// If the type doesn’t have a MarshalJSON() method, then Go will fall back to trying to encode
// it to JSON based on its own internal set of rules.
func (r Runtime) MarshalJSON()([]byte, error) {
	// concatinate min with the value
	jsonValue := fmt.Sprintf("%d min",r)
	// adding double quotes to make it as json format
	quotesjsonvalue := strconv.Quote(jsonValue)
	
	return []byte(quotesjsonvalue),nil
} 


//type UnMarshaler interface {
// UnMarshalJSON() ([]byte, error)
// }
// If the type does satisfy the interface, then Go will call its UnMarshalJSON() method and use the
// []byte slice that it returns as the encoded JSON value.
// If the type doesn’t have a UnMarshalJSON() method, then Go will fall back to trying to decode
// it to JSON based on its own internal set of rules.


func (r *Runtime)UnmarshalJSON(jsonValue []byte) error {
	// unqoute the json value "value" -> value
	unquotedJSONValue,err := strconv.Unquote(string(jsonValue)) 
	if err!=nil {
		return ErrInvalidRuntimeFormat
	}

	parts := strings.Split(unquotedJSONValue, " ")


	// parts is a []string so if is length is not 2 that means its not the runtime format and parts[1] is the min value
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	} 

	// convert the integer32 with base 10  part[0] = "100" -> part[0] = 100 part[1] = "mins"
	i , err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// type conversion to Runtime
	*r = Runtime(i)

	return  nil
}