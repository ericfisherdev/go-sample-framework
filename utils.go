package finishline

import (
	"fmt"
	"regexp"
	"runtime"
	"time"
)

func (f *FinishLine) LoadTime(start time.Time) {
	elapsed := time.Since(start)
	caller, _, _, _ := runtime.Caller(1)
	funcObj := runtime.FuncForPC(caller)
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	f.InfoLog.Println(fmt.Sprintf("Load Time: %s took %s", name, elapsed))
}
