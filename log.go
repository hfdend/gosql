package gosql

import (
    "io"
    "fmt"
    "time"
    "sync"
)

var (
    logWriter   io.Writer
    green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
    white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
    yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
    red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
    blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
    magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
    cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
    reset   = string([]byte{27, 91, 48, 109})
    logMutex sync.Mutex
)

func LogSetOutput(writer io.Writer) {
    logMutex.Lock()
    defer logMutex.Unlock()
    logWriter = writer
}

func logWrite(sql ExecSql) {
    if logWriter == nil {
        return
    }
    logMutex.Lock()
    defer logMutex.Unlock()
    fmt.Fprintln(logWriter, "[gosql]", time.Now().Format("2006/01/02 15:04:05"), sql)
}

