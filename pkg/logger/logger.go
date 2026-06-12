package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/clitorhea/rhea-note/pkg/config"
)

var (
	debugLogger *log.Logger
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugMode   bool
)

func Init(debug bool) error {
	debugMode = debug
	
	logDir := config.ConfigDir()
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return err
	}
	
	logPath := filepath.Join(logDir, "secnotes.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	debugLogger = log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(file, "INFO:  ", log.Ldate|log.Ltime)
	errorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

func Debugf(format string, v ...interface{}) {
	if debugMode && debugLogger != nil {
		debugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func Infof(format string, v ...interface{}) {
	if infoLogger != nil {
		infoLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func Errorf(format string, v ...interface{}) {
	if errorLogger != nil {
		errorLogger.Output(2, fmt.Sprintf(format, v...))
	}
}
