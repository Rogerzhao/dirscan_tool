package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/Rogerzhao/xmlib/xmlog"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type DirScanner struct {
	path       string
	dirFilter  string
	dirRegexp  *regexp.Regexp
	fileFilter string
	fileRegexp *regexp.Regexp
	resultFile string
	fileInfos  []*FileInfo

	failFilterFiles []string
}

type FileInfo struct {
	fileName string
	size     int64
	sha1     string
}

func (f *FileInfo) genSha1() (err error) {
	buf, err := ioutil.ReadFile(f.fileName)
	if err != nil {
		xmlog.ERROR(err)
		return
	}

	h := sha1.New()
	io.WriteString(h, string(buf))
	f.sha1 = fmt.Sprintf("%x", h.Sum(nil))
	output := fmt.Sprintf("%s,%s,%d", f.fileName, f.sha1, f.size)
	outputChan <- output
	syncChan <- 1

	concurrentChan <- 1
	return
}

func NewDirScanner(filePath string, dirFilter string, fileFilter string, resultFile string) (d *DirScanner, err error) {
	d = &DirScanner{
		path:       filePath,
		dirFilter:  dirFilter,
		fileFilter: fileFilter,
		resultFile: resultFile,
	}
	err = d.CompileRegexp()
	if err != nil {
		xmlog.ERROR(err)
		return
	}
	return
}

func (d *DirScanner) scanDirWithFilter() {
	xmlog.Infof("start to scan file info ")
	filepath.Walk(d.path, d.visit)
}

func (d *DirScanner) ScanFileInfo() {
	d.scanDirWithFilter()
	var count int64
	count = 0
	for _, fileInfo := range d.fileInfos {
		go fileInfo.genSha1()
		count++
		<-concurrentChan
	}

	// sync
	for i := int64(0); i < count; i++ {
		<-syncChan
	}
	close(outputChan)
	xmlog.Infof("end scan file info ,get %d files", count)
}

func (d *DirScanner) CompileRegexp() (err error) {
	d.dirFilter = strings.Replace(d.dirFilter, "!", "^", -1)
	d.dirFilter = strings.Replace(d.dirFilter, "?", `[a-zA-Z0-9\_\-\.]{1}`, -1)
	d.dirFilter = strings.Replace(d.dirFilter, "*", `[a-zA-Z0-9\_\-\.]*`, -1)
	d.dirFilter = fmt.Sprintf("^%s$", d.dirFilter)
	xmlog.Infof("dirFilter %s", d.dirFilter)
	d.dirRegexp, err = regexp.Compile(d.dirFilter)
	if err != nil {
		xmlog.ERROR(err)
		return
	}

	d.fileFilter = strings.Replace(d.fileFilter, "!", "^", -1)
	d.fileFilter = strings.Replace(d.fileFilter, "?", `[a-zA-Z0-9\_\-\.]{1}`, -1)
	d.fileFilter = strings.Replace(d.fileFilter, "*", `[a-zA-Z0-9\_\-\.]*`, -1)
	d.fileFilter = fmt.Sprintf("^%s$", d.fileFilter)
	xmlog.Infof("fileFilter %s", d.fileFilter)
	d.fileRegexp, err = regexp.Compile(d.fileFilter)
	if err != nil {
		xmlog.ERROR(err)
		return
	}
	return
}

func (d *DirScanner) visit(path string, f os.FileInfo, err1 error) (err error) {
	if err1 != nil {
		err = err1
		xmlog.ERROR(err)
		return
	}
	if f == nil {
		return
	}
	if f.IsDir() {
		return
	}

	dirArr := strings.Split(path, "/")
	fileName := dirArr[len(dirArr)-1]
	dirArr = dirArr[:len(dirArr)-1]
	// check dir
	for _, dirName := range dirArr {
		if d.dirRegexp.MatchString(dirName) {
			xmlog.Infof("dir %s match %s", dirName, d.dirFilter)
			d.failFilterFiles = append(d.failFilterFiles, path)
			return
		}
	}

	// check fileName
	if d.fileRegexp.MatchString(fileName) {
		xmlog.Infof("filename  %s match %s", fileName, d.fileFilter)
		d.failFilterFiles = append(d.failFilterFiles, path)
		return
	}
	xmlog.Infof("get file %s ", path)
	info := &FileInfo{
		fileName: path,
		size:     f.Size(),
	}
	d.fileInfos = append(d.fileInfos, info)
	return

}

func (d *DirScanner) fileStore() (err error) {
	file, err := os.OpenFile(d.resultFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		xmlog.ERROR(err)
		return
	}
	defer file.Close()
	for line := range outputChan {
		_, err = file.WriteString(line + "\n")
		if err != nil {
			xmlog.ERROR(err)
			continue
		}

	}
	quitChan <- 1
	return

}
