package utils

import (
	"runtime"
	"strings"
)

const (
	depth = 1
)

// GetCurrentFunctionName возвращает имя функции/метода,
// из которой была вызвана
func GetCurrentFunctionName() string {
	function, _, _, ok := runtime.Caller(depth)
	if !ok {
		return "Unknown function"
	}

	fullName := runtime.FuncForPC(function).Name()

	lastSlash := strings.LastIndex(fullName, "/")
	if lastSlash != -1 {
		return fullName[lastSlash+1:]
	}

	return fullName
}
