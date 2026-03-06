package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)


type envelope map[string]any

func (app *application) readIDParams(r *http.Request) (int64, error) {
	// returns a params slice(with elements as struct) with |key1,value1|key2,value2|...
	params := httprouter.ParamsFromContext(r.Context())
	//.parseint that can be used to convert string into int with wide varity of bases like 0,8,16,32,64,we can also use atoi
	id, err := strconv.ParseInt(params.ByName("id"),10,64)
	if err!=nil || id < 1{
		return 0, errors.New("Invalid ID Parameter")
	}
	return id,nil
}

func (app *application) writeJSON(data envelope,w http.ResponseWriter, status int, headers http.Header)  error{
	//marshall return a byte array and an error , MarshalIndent , no line prefix and tab indent for each element
	js, err := json.Marshal(data)
	if err!=nil {
		return err
	}

	for key,value:= range headers {
		// just sending other headers
		w.Header()[key] = value
	}
	// for looking better in terminal application...
	js = append(js, '\n')
	// sending the json type for notify the browser the content is json
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}