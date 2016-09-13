package xmlog

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// Xmlogger 日志定义
type Xmlogger struct {
	prefix   string
	errChan  chan error
	quitChan chan bool
	start    bool
	hookfunc func(err error)
}

// SetHookFunc 设置hook函数
func (l *Xmlogger) SetHookFunc(f func(err error)) {
	l.hookfunc = f
}

func (l *Xmlogger) printf(format string, args ...interface{}) error {
	return l.wrapError(fmt.Errorf(format, args...), 3)
}
func (l *Xmlogger) print(err error) error {
	return l.wrapError(err, 3)
}

// Printf 格式化封装args
func (l *Xmlogger) Printf(format string, args ...interface{}) error {
	return l.wrapError(fmt.Errorf(format, args...), 2)
}

// Print 直接封装
func (l *Xmlogger) Print(err error) error {
	return l.wrapError(err, 2)
}

// PrintN 指定封装层次封装
func (l *Xmlogger) PrintN(depth int, err error) error {
	return l.wrapError(err, depth)
}

// PrintfN 格式化的指定层次封装
func (l *Xmlogger) PrintfN(depth int, format string, args ...interface{}) error {
	return l.wrapError(fmt.Errorf(format, args...), depth)
}

//Close 关闭文件
func (l *Xmlogger) Close() {
	if !l.Start() {
		return
	}
	close(l.errChan)
	<-l.quitChan
}

// Start 文件写入是否已经开始
func (l *Xmlogger) Start() bool {
	return l.start
}

// NewLogger 使用文件前缀初始化
func NewLogger(prefix string) *Xmlogger {
	return &Xmlogger{
		prefix:   prefix,
		errChan:  make(chan error, 1024),
		quitChan: make(chan bool),
		start:    false,
	}
}

// NewLoggerN 使用文件前缀，errChan,quitChan初始化
func NewLoggerN(prefix string, errChan chan error, errQuit chan bool) *Xmlogger {
	return &Xmlogger{
		prefix:   prefix,
		errChan:  errChan,
		quitChan: errQuit,
		start:    false,
	}
}

// WrapError(err) -> caller(1)
// WrapError(err, n) -> caller(n)
func (l *Xmlogger) wrapError(err error, trackStack int) (nerr error) {

	_, file, line, _ := runtime.Caller(trackStack)
	file = filepath.Base(file)
	var errMsg string
	if l.prefix != "" {
		errMsg = fmt.Sprintf("%s:%d %s %s", file, line, l.prefix, err.Error())
	} else {
		errMsg = fmt.Sprintf("%s:%d %s", file, line, err.Error())
	}
	nerr = errors.New(errMsg)
	l.errChan <- nerr
	return
}

// WatchErrors 启动错误日志监控
// 参数 prefix 日志前缀 logDir 日志目录
func (l *Xmlogger) WatchErrors(prefix string, logdir string) {
	if l.start {
		return
	}
	var now = time.Now()
	var prevYear, prevDay int
	var curYear, curDay int
	var prevMonth time.Month
	var curMonth time.Month

	logdir = strings.TrimRight(logdir, "/")
	logFilename, baseFilename := getCurrLogName(logdir, prefix, now)
	symblink := l.getSymbname(logdir, prefix)
	os.Remove(symblink)
	os.Symlink(baseFilename, symblink)
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Create log file fail: %s\n", err)
		return
	}
	l.start = true
	prevYear, prevMonth, prevDay = now.Year(), now.Month(), now.Day()
	for err := range l.errChan {
		now = time.Now()
		curYear = now.Year()
		curMonth = now.Month()
		curDay = now.Day()
		if prevYear != curYear || prevMonth != curMonth || prevDay != curDay {
			logFilename, baseFilename = getCurrLogName(logdir, prefix, now)
			newlogFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintf(logFile, "open new log file %s failed: %s\n", logFilename, err)
			} else {
				logFile.Close()
				logFile = newlogFile
				symblink := l.getSymbname(logdir, prefix)
				os.Remove(symblink)
				os.Symlink(baseFilename, symblink)
			}
		}
		prevYear, prevMonth, prevDay = curYear, curMonth, curDay
		err = fmt.Errorf("%d-%02d-%02d %d:%02d:%02d %s",
			curYear, curMonth, curDay, now.Hour(), now.Minute(), now.Second(),
			err.Error())
		fmt.Fprintf(logFile, "%s\n", err)
		if l.hookfunc != nil {
			l.hookfunc(err)
		}
	}
	logFile.Close()
	l.quitChan <- true
}

func getCurrLogName(logdir string, prefix string, now time.Time) (fullFilename, baseFilename string) {
	baseFilename = fmt.Sprintf("%s-%d%02d%02d.log", prefix, now.Year(), now.Month(), now.Day())
	fullFilename = path.Join(logdir, baseFilename)
	return
}

func (l *Xmlogger) getSymbname(logdir, prefix string) (symblink string) {
	filename := fmt.Sprintf("%s-current.log", prefix)
	return path.Join(logdir, filename)
}

// DumpStack 输出错误栈到文件
func (l *Xmlogger) DumpStack() {
	cnt := 1
	l.Printf("------- DumpStack --------")
	for {
		_, file, line, ok := runtime.Caller(cnt)
		if !ok {
			break
		}
		l.Printf("%s:%d", file, line)
		cnt++
	}
}

// PanicDumpStack 输出Panic栈到文件
func (l *Xmlogger) PanicDumpStack(err interface{}) {
	cnt := 1
	l.Printf("------- PANIC %v --------", err)
	for {
		_, file, line, ok := runtime.Caller(cnt)
		if !ok {
			break
		}
		l.Printf("%s:%d", file, line)
		cnt++
	}
}

// HandlePanic 处理panic
func (l *Xmlogger) HandlePanic() {
	if err := recover(); err != nil {
		l.PanicDumpStack(err)
	}
}

// HandlePanic 处理抛出的panic
func HandlePanic() {
	DefaultLog.HandlePanic()
}

// RedirectToPanicFile 将panic信息打到.panic
func RedirectToPanicFile() {
	var discard *os.File
	var err error
	discard, err = os.OpenFile(".panic", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		discard, err = os.OpenFile("/dev/null", os.O_RDWR, 0)
	}
	if err == nil {
		fd := discard.Fd()
		syscall.Dup2(int(fd), int(os.Stderr.Fd()))
	}
}

// DumpStack 默认的DumpStack
func DumpStack() {
	DefaultLog.DumpStack()
}

// PanicDumpStack 默认的PanicDumpStack
func PanicDumpStack(err interface{}) {
	DefaultLog.PanicDumpStack(err)
}

// WatchErrors 默认的WatchErrors
func WatchErrors(prefix string, logDir string) {
	go DefaultLog.WatchErrors(prefix, logDir)
}

// Watch 默认的WatchErrors
func Watch(prefix string, logDir string) {
	go DefaultLog.WatchErrors(prefix, logDir)
}

var errChan = make(chan error, 1024)
var quitChan = make(chan bool)

// DefaultLog 默认的日志对象
var DefaultLog = NewLoggerN("", errChan, quitChan)

// Close 关闭默认的日志文件
// close will close xmlogger or synclogger
func Close() {
	if DefaultLog.Start() {
		DefaultLog.Close()
	}
}

// ERROR 输出错误日志信息
func ERROR(err error) error {
	err = fmt.Errorf("ERROR %s", err)
	return DefaultLog.print(err)
}

// Error 输出错误日志
func Error(err error) error {
	err = fmt.Errorf("ERROR %s", err)
	return DefaultLog.print(err)
}

// ERRORF 格式化输出错误信息
func ERRORF(format string, args ...interface{}) error {
	format = "ERROR " + format
	return DefaultLog.print(fmt.Errorf(format, args...))
}

// Errorf 格式化输出错误信息
func Errorf(format string, args ...interface{}) error {
	format = "ERROR " + format
	return DefaultLog.print(fmt.Errorf(format, args...))
}

// DEBUG 输出DEBUG信息
func DEBUG(err error) error {
	err = fmt.Errorf("DEBUG %s", err)
	return DefaultLog.print(err)
}

// Debug 输出DEBUG信息
func Debug(err error) error {
	err = fmt.Errorf("DEBUG %s", err)
	return DefaultLog.print(err)
}

// DEBUGF 格式化输出DEBUG信息
func DEBUGF(format string, args ...interface{}) error {
	format = "DEBUG " + format
	return DefaultLog.print(fmt.Errorf(format, args...))
}

// Debugf 格式化输出DEBUG信息
func Debugf(format string, args ...interface{}) error {
	format = "DEBUG " + format
	return DefaultLog.print(fmt.Errorf(format, args...))
}

// INFO 输出INFO信息
func INFO(err error) error {
	err = fmt.Errorf("INFO %s", err)
	return DefaultLog.print(err)
}

// Info 输出INFO信息
func Info(err error) error {
	err = fmt.Errorf("INFO %s", err)
	return DefaultLog.print(err)
}

// INFOF 格式化输出INFO信息
func INFOF(format string, args ...interface{}) error {
	format = "INFO " + format
	return DefaultLog.print(fmt.Errorf(format, args...))
}

// Infof 格式化输出INFO信息
func Infof(format string, args ...interface{}) error {
	format = "INFO " + format
	return DefaultLog.print(fmt.Errorf(format, args...))
}

// WARN 输出WARN信息
func WARN(err error) error {
	err = fmt.Errorf("WARN %s", err)
	return DefaultLog.print(err)
}

// Warn 输出WARN信息
func Warn(err error) error {
	err = fmt.Errorf("WARN %s", err)
	return DefaultLog.print(err)
}

// WARNF 格式化输出WARN信息
func WARNF(format string, args ...interface{}) error {
	format = "WARN " + format
	return DefaultLog.print(fmt.Errorf(format, args...))
}

// Warnf 格式化输出WARN信息
func Warnf(format string, args ...interface{}) error {
	format = "WARN " + format
	return DefaultLog.print(fmt.Errorf(format, args...))
}
