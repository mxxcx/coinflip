package logger

import (
    "io/ioutil"
    "log"
	"os"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
    Trace   *log.Logger
    Info    *log.Logger
    Warning *log.Logger
    Error   *log.Logger
)

// Init ...
func init() {
    Trace = log.New(ioutil.Discard, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
    Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
    Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
		
	Info.SetOutput(&lumberjack.Logger{
		Filename:   "logs/info/info.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Compress:   true, // disabled by default
	})	
	Warning.SetOutput(&lumberjack.Logger{
		Filename:   "logs/warning/warning.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Compress:   true, // disabled by default
	})
	Error.SetOutput(&lumberjack.Logger{
		Filename:   "logs/error/error.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     56, //days
		Compress:   true, // disabled by default
	})
}