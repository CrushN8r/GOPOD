package logger

import "fmt"

func LogDebugUI(a ...any) {
	LogTray(a...)
}

// LogUIPlain/LogDebugUIPlain: same storage as LogUI/LogDebugUI above, but
// without the timestamp prefix - for the one BROBOT_RICH_DISPLAY line in
// sdkapp/server.go that wants just "Brobot 1: <text>", not every other
// wire-pod log line (which keeps LogUI/LogDebugUI unchanged).
func LogUIPlain(a ...any) {
	LogArray = append(LogArray, fmt.Sprint(a...)+"\n")
	if len(LogArray) >= 50 {
		LogArray = LogArray[1:]
	}
	LogList = ""
	for _, b := range LogArray {
		LogList = LogList + b
	}
}

func LogDebugUIPlain(a ...any) {
	LogTrayPlain(a...)
}

func LogTrayPlain(a ...any) {
	LogTrayArray = append(LogTrayArray, fmt.Sprint(a...)+"\n")
	if len(LogTrayArray) >= 200 {
		LogTrayArray = LogTrayArray[1:]
	}
	LogTrayList = ""
	for _, b := range LogTrayArray {
		LogTrayList = LogTrayList + b
	}
	select {
	case LogTrayChan <- fmt.Sprint(a...) + "\n":
	default:
	}
}
