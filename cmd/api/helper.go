package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// for json reponse making use json encoding and json marshal()
// json marshal() use more memory and can be customized
// json encode use less memory and high performace but customization is not possible(not so flexible)




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

// json.decode only work once that means the first json is readed , if we again call json.decode then it reads the second json


func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dsn any) error {
	// this prevent ddos attacks and sending malicious code from client ...
	// _ is something numeric literal separator if we use 1000000 is 10lak == 1_000_000

	maxBytes :=1_048_576
	// http.maxbytereader is like a security guard, tries to read the body
	//  (e.g., using io.ReadAll or a JSON decoder), MaxBytesReader tracks the bytes:
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes)) // convert the no. into int64

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // this will check the disallowed field when encode
	err := dec.Decode(&dsn)

	if err!=nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxByteError *http.MaxBytesError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)",syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err,&unmarshalTypeError):
			if unmarshalTypeError.Field != ""{
				return fmt.Errorf("body contains incorrect JSON type for field %q",unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// Does this error message start with these exact words
		case strings.HasPrefix(err.Error(), "json: unknown field ") :
			fieldname := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return  fmt.Errorf("body contains unknown key %s", fieldname)

		case errors.As(err,&maxByteError):
			return fmt.Errorf("body must not be larger than %d bytes",maxByteError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
		

	}
	// &struct{}{} (an empty anonymous struct) because we don't actually care about the data
	err = dec.Decode(&struct{}{})
	if err!= io.EOF {   // this check the io.EOF (end of the read source) 
	// so if err is  nil decoder found some more json data 
		return errors.New("body must only contain a single JSON value")
	}
	return  nil
}

func (app *application)badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	// we use err.Error() because err is a interface we need to call the function to get the err string
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}