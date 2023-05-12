package logger

import (
	"fmt"
	"time"
)

func getTime() string {
	dt := time.Now()
	return dt.Format(time.TimeOnly)
}

func Info(v ...any) {
	fmt.Println(getTime() + " [\x1b[36mINFO\x1b[0m] " + fmt.Sprint(v...))
}

func Error(v ...any) {
	fmt.Println(getTime() + " [\x1b[31mERROR\x1b[0m] " + fmt.Sprint(v...))
}

func Warn(v ...any) {
	fmt.Println(getTime() + " [\x1b[33mWARN\x1b[0m] " + fmt.Sprint(v...))
}
