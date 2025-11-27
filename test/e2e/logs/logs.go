package logs

import (
	"fmt"
	"log"
	"time"

	"github.com/onsi/ginkgo/v2"
	"k8s.io/klog/v2"
)

const defaultFlush = 5 * time.Second

type klogBridge struct{}

func (klogBridge) Write(p []byte) (int, error) {
	klog.InfoDepth(1, string(p))
	return len(p), nil
}

func Init() {
	log.SetOutput(klogBridge{})
	log.SetFlags(0)
	klog.StartFlushDaemon(defaultFlush)
	klog.EnableContextualLogging(false)
}

func Close() {
	klog.Flush()
}

func stamp() string {
	return time.Now().Format(time.StampMilli) + ": "
}

func writef(level, format string, args ...interface{}) {
	if _, err := fmt.Fprintf(ginkgo.GinkgoWriter, "%s%s: %s\n", stamp(), level, fmt.Sprintf(format, args...)); err != nil {
		klog.Errorf("failed to write ginkgo log: %v", err)
	}
}

func Infof(format string, args ...interface{}) {
	writef("INFO", format, args...)
	klog.InfoDepth(1, fmt.Sprintf(format, args...))
}

func Debugf(v int32, format string, args ...interface{}) {
	klog.V(klog.Level(v)).InfoDepth(1, fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}) {
	writef("WARN", format, args...)
	klog.WarningDepth(1, fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}) {
	writef("ERROR", format, args...)
	klog.ErrorDepth(1, fmt.Sprintf(format, args...))
}

func Stepf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	ginkgo.By(msg)
	writef("STEP", "%s", msg)
}

func Failf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	writef("FAIL", "%s", msg)
	const skip = 1
	ginkgo.Fail(msg, skip)
	panic("unreachable")
}

var Fail = ginkgo.Fail
