package data

import (
	"fmt"
	"strconv"
)

type Runtime int32


func (r Runtime) MarshalJSON()([]byte, error) {
	// concatinate min with the value
	jsonValue := fmt.Sprintf("%d min",r)
	// adding double quotes to make it as json format
	quotesjsonvalue := strconv.Quote(jsonValue)
	
	return []byte(quotesjsonvalue),nil
} 
