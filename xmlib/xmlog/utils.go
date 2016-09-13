package xmlog

// write by zhangye@xiaomi.com

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// WritePid 写入pid
func (l *Xmlogger) WritePid(listenPort string) (err error) {
	pid := strconv.Itoa(os.Getpid())
	if pid == "" {
		return
	}
	filename := "web." + strings.Split(listenPort, ":")[1] + ".pid"
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer func() {
		file.Close()
	}()
	file.WriteString(pid)

	return
}

// WatchPanic panic重定向到文件
func (l *Xmlogger) WatchPanic() {
	var panicFile *os.File
	program := filepath.Base(os.Args[0])
	filePath := fmt.Sprintf("/tmp/%s_%d.panic", program, os.Getpid())
	panicFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panicFile, err = os.OpenFile("/dev/null", os.O_RDWR, 0)
	}
	if err == nil {
		fd := panicFile.Fd()
		syscall.Dup2(int(fd), int(os.Stderr.Fd()))
	}
}

// WritePid 将文件pid输出到以prefix，listenPort命名的文件中
func WritePid(prefix string, listenPort string) (err error) {
	pid := strconv.Itoa(os.Getpid())
	if pid == "" {
		return
	}
	if strings.Contains(listenPort, ":") {
		listenPort = strings.Split(listenPort, ":")[1]
	}
	filename := fmt.Sprintf("%s.%s.pid", prefix, listenPort)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer func() {
		file.Close()
	}()
	file.WriteString(pid)

	return
}

// WritePidFile 将文件pid输出到以prefix，listenPort命名的文件中
func WritePidFile(prefix string) (err error) {
	filename := fmt.Sprintf("%s.pid", prefix)
	err = ioutil.WriteFile(filename, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)

	return
}

// WatchPanic 默认的panic重定向函数
func WatchPanic() {
	var panicFile *os.File
	panicFile, err := os.OpenFile(".panic", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panicFile, err = os.OpenFile("/dev/null", os.O_RDWR, 0)
	}
	if err == nil {
		fd := panicFile.Fd()
		syscall.Dup2(int(fd), int(os.Stderr.Fd()))
	}
}

// GetFunctionName get current function name
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// Pretty pretty print any thing into json
func Pretty(writer io.Writer, obj interface{}) (err error) {
	content, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(writer, "%s", content)
	return
}
