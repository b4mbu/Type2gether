package logger

import (
	"fmt"
    "io"
	"sync"
	"time"

	"github.com/fatih/color"
)


type Logger struct {
    writerMutex  sync.Mutex
    writer io.Writer
}

func NewLogger(w io.Writer) *Logger {
    return &Logger{writer: w}
}

func (log *Logger) Info(message string) {
    blue := color.New(color.FgBlue).FprintlnFunc()
    log.printMessage(blue, message)
}

func (log *Logger) Success(message string) {
    green := color.New(color.FgGreen).FprintlnFunc()
    log.printMessage(green, message)
}

func (log *Logger) Error(message string) {
    red := color.New(color.FgRed).FprintlnFunc()
    log.printMessage(red, message)
} 

func (log *Logger) printMessage(colorPrint func(w  io.Writer, a ...interface{}), message string) {
    log.writerMutex.Lock()
    defer log.writerMutex.Unlock()

    colorPrint(log.writer, fmt.Sprintf("[%s] ", time.Now().Format("2006-01-02 15:04:05")) + message)
}

