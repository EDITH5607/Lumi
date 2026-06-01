package main

import (
	"Green/internal/validator"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// for json reponse making use json encoding and json marshal()
// json marshal() use more memory and can be customized
// json encode use less memory and high performace but customization is not possible(not so flexible)




type envelope map[string]any



//read id query parameter from the url
func (app *application) readIDParams(r *http.Request) (int64, error) {

	// returns a params slice(with elements as struct) with |key1,value1|key2,value2|...
	params := httprouter.ParamsFromContext(r.Context())

	//.parseint that can be used to convert string into int with wide varity of bases like 0,8,16,32,64,we can also use atoi
	id, err := strconv.ParseInt(params.ByName("id"),10,64)  // the id:value , parse value into int 
	if err!=nil || id < 1{
		return 0, errors.New("Invalid ID Parameter")
	}
	return id,nil
}



// write json used to write content as json using marshal method
func (app *application) writeJSON(data envelope,w http.ResponseWriter, status int, headers http.Header)  error{
	//marshall return a byte array and an error , MarshalIndent , no line prefix and tab indent for each element
	js, err := json.Marshal(data)
	if err!=nil {
		return err
	}

	// for looking better in terminal application...
	js = append(js, '\n')
	for key,value:= range headers {
		// just sending other headers
		w.Header()[key] = value
	}
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


func (app *application) readString(qs url.Values, key string, defaultString string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultString
	}
	return s
}

func (app *application) readInt(qs url.Values,key string, defaultvalue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s==""{
		return defaultvalue
	}
	
	i, err := strconv.Atoi(s)
	if err!=nil {
		v.AddError(key, "must be an integer value")
		return defaultvalue
	}
	return i
}


// the data look like "key":"value1,value2,value3"->get()->"value1,value2,value3"
func (app *application) readCSV(qs url.Values, key string, defaultvalue []string) []string {
	csv := qs.Get(key)
	if csv=="" {
		return defaultvalue
	}
	return strings.Split(csv,",") // split fn return a string[] with split values
}


// background is a goroutine factory which accept fn as parameter 
// run the fn inside a goroutine safely if any panic occur it catches and recover that goroutine
// without effecting the whole program
func (app *application) background(fn func()) {

	// launching background goroutine
	// if a goroutine panics (crashes) it will take down entire program, even if main programs is fine
	// defer runs after the surrounding function ends — whether it ends normally or by panic.
	// we make it as a defer fn because a block of code is need to run its not a goroutine
	// err handles expected errors. But some errors are unexpected panics (like nil pointer, index out of range).
	//  recover() catches those panics that err would never catch.


	// increment the waitgroup counter with 1 
	app.wg.Add(1)
	go func ()  {

		// ,done() decrement the waitgroup counter
		// first func() defer execute and then defer app.wg... execute means wg decrement at the last
		defer app.wg.Done()
		// Recover any panic
		defer func() {
			if err := recover(); err!=nil {
				stack := debug.Stack()
        			app.logger.PrintError(fmt.Errorf("panic: %v\n%s", err, stack), nil)			}
		}()	

		// Execute the arbitary fn that we passed  as parameter
		fn()
	}()
}