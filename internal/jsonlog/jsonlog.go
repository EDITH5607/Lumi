package jsonlog

import (
	"encoding/json"
	"io"
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

func (l *Logger) PrintError(message string, properties map[string]string) {
	l.print(LevelError, message, properties)
}


//this print is private because fn start with lowercase letter
func (l *Logger) print(level Level, message string, properties map[string]string) (int,error) {
	//checking the severirty of the log entry
	if level < l.minLevel {
		return 0, nil
	}
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
	if level>=LevelError {
		aux.Trace = string(debug.Stack())
	}

	var line []byte
	line,err := json.Marshal(aux)
	if err!=nil {
		
	}
	return  0, nil
}