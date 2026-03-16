package data

import (
	"fmt"
	"strconv"
)

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
