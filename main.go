package main

import (
	"flag"
	"fmt"
	"github.com/Rogerzhao/xmlib/config"
	"github.com/Rogerzhao/xmlib/xmlog"
	"os"
	"strings"
)

var (
	gConf   config.Configer
	cfgFile = flag.String("c", "", "config file")
	// for log
	logPath string
	prefix  string

	// for sha1 scan
	outputChan     = make(chan string, 100)
	syncChan       = make(chan int, 100)
	quitChan       = make(chan int)
	walkPath       string
	filterDirName  string
	filterFileName string
	resultFile     string

	// concurrentFileNumber
	concurrentNumber int64
	concurrentChan   chan int
)

func main() {
	flag.Parse()
	if *cfgFile == "" {
		fmt.Printf("Usage: %s -c=etc/gamepackage.conf", os.Args[0])
		os.Exit(1)
	}

	err := Init(*cfgFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	concurrentChan = make(chan int, concurrentNumber)

	defer func() {
		xmlog.Close()
	}()
	go xmlog.WatchErrors(prefix, logPath)
	go xmlog.WatchPanic()
	dirScanner, err := NewDirScanner(walkPath, filterDirName, filterFileName, resultFile)
	if err != nil {
		xmlog.ERROR(err)
		return
	}
	go dirScanner.fileStore()
	// 获取文件列表
	dirScanner.ScanFileInfo()
	<-quitChan
	xmlog.Infof("scan dir successfully end")
	// 生成文件
}

func Init(fileName string) (err error) {
	gConf, err = config.NewConfiger(fileName)
	if err != nil {
		return
	}
	logPath, err = gConf.GetSetting("log_conf", "logdir")
	if err != nil {
		return err
	}
	prefix, err = gConf.GetSetting("log_conf", "prefix")
	if err != nil {
		return err
	}
	walkPath, err = gConf.GetSetting("path", "walkPath")
	if err != nil {
		return
	}
	// currentFile
	if strings.Contains(walkPath, `~`) {
		home := os.Getenv("HOME")
		if len(home) > 0 {
			walkPath = strings.Replace(walkPath, `~`, home, -1)
		}
	}
	fmt.Println(walkPath)

	filterDirName, err = gConf.GetSetting("path", "filterDir")
	if err != nil {
		return
	}
	filterFileName, err = gConf.GetSetting("path", "filterFile")
	if err != nil {
		return
	}
	concurrentNumber, err = gConf.GetIntSetting("path", "concurrentNumber", 64)
	if err != nil {
		return
	}
	resultFile, err = gConf.GetSetting("path", "resultFile")
	if err != nil {
		return
	}
	return
}
