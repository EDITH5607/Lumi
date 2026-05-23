package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type Level int


// iota wil automatically assign values to the constants like Levelinfo=0 and LevelError =1  and so on
// the LevelError, LevelFatal, LevelOff will be the same type Level 
const(
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)


func(l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return  "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""	
	}
}

type Logger struct {
	out io.Writer
	minLevel Level
	mu sync.Mutex
}

func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out: out,
		minLevel: minLevel,
	}
}

func (l *Logger)PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1)
}


//this print is private because fn start with lowercase letter
func (l *Logger) print(level Level, message string, properties map[string]string) (int,error) {

	//checking the severirty of the log entry
	if level < l.minLevel {
		return 0, nil
	}
	//build a structured log entry
	aux := struct {
		Level		string		  `json:"level"`
		Time  	string		  `json:"time"`
		Message 	string		  `json:"message"`
		Properties  map[string]string   `json:"properties,omitempty"`
		Trace       string		  `json:"trace,omitempty"`
	} {
		Level: level.String(),
		Time: time.Now().UTC().Format(time.RFC3339),
		Message: message,
		Properties: properties,
	}

	//attach stack trace for errors
	if level>=LevelError {
		aux.Trace = string(debug.Stack())
	}

	// Marshal to json...
	var line []byte
	line,err := json.Marshal(aux)
	if err!=nil {
		line = []byte(LevelError.String()+":unable to marshal log message:"+ err.Error())
	}

	//lock to avoid concurrent write
	l.mu.Lock() // mutex lock to lock to avoid two writes  output ,destination can happen concurrently
	defer l.mu.Unlock() // run after return 
	return  l.out.Write(append(line,'\n'))  // write  like io.write([]byte(data))  is like os.Stdout.Write() or file.Write() not the below write
}


 //by implementing Write(), the Logger itself becomes an io.Writer. 
 // This means you can pass it anywhere Go expects an io.Writer, like http.Server's error logger:
 //  srv := &http.Server{
//     ErrorLog: log.New(logger, "", 0), // uses your custom logger
// }
func (l *Logger) Write(message []byte)(n int, err error) {
	return l.print(LevelError, string(message), nil)
}