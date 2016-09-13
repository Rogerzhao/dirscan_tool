package main

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
	"xmlib/xmlog"
)

func Test_filterStragety(t *testing.T) {
	go xmlog.WatchErrors("test", "log")
	go xmlog.WatchPanic()

	defer xmlog.Close()
	dirFilter := "*e?[!a-z]*"
	fileFilter := "*e?[!a-z]*"
	p, _ := NewDirScanner("/tmp/api_tantan", dirFilter, fileFilter, "/tmp/sha1.output")
	p.scanDirWithFilter()
	fmt.Printf("%s\n", p.dirFilter)
	regexpStr := `[a-zA-Z0-9\.\-\_]*e[a-zA-Z0-9\.\-\_]{1}[^a-z][a-zA-Z0-9\.\-\_]*`
	fmt.Printf("%s\n", regexpStr)
	reg, err := regexp.Compile(regexpStr)
	if err != nil {
		t.Fatal(err)
	}
	for _, fileName := range p.failFilterFiles {
		fmt.Println(fileName + "  skip")
		path := fileName[:strings.LastIndex(fileName, "/")]
		name := fileName[strings.LastIndex(fileName, "/")+1:]
		if !reg.MatchString(path) && !reg.MatchString(name) {
			t.Fatalf("fileName should not be filtered %s", fileName)
		}

	}
}

func Test_sha1Result(t *testing.T) {
	go xmlog.WatchErrors("test", "log")
	go xmlog.WatchPanic()

	defer xmlog.Close()
	dirFilter := "*e?[!a-z]*"
	fileFilter := "*e?[!a-z]*"
	syncChan = make(chan int, 100) //
	p, _ := NewDirScanner("/tmp/api_tantan", dirFilter, fileFilter, "/tmp/sha1.output")
	go p.fileStore()
	p.ScanFileInfo()
	<-quitChan

	fileNumber := len(p.fileInfos)
	file, err := os.Open("/tmp/sha1.output")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	readBuffer := bufio.NewReader(file)
	lineCount := 0
	// check the first 10 line for sha1
	for {
		line, err := readBuffer.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if lineCount < 10 {
			ok, err := checkSha1(line)
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatalf("sha1 check fail for %s", line)
			}
		}
		lineCount++
	}

	if lineCount != fileNumber {
		t.Fatalf("sha1 lines missing expect %d lines but get %d lines", fileNumber, lineCount)
	}

}

func checkSha1(line string) (result bool, err error) {
	lineArr := strings.Split(line, ",")
	if len(lineArr) != 3 {
		err = fmt.Errorf("line format error, expect 3 columns")
		return
	}
	buf, err := ioutil.ReadFile(lineArr[0])
	if err != nil {
		xmlog.ERROR(err)
		return
	}

	h := sha1.New()
	io.WriteString(h, string(buf))
	sha1 := fmt.Sprintf("%x", h.Sum(nil))
	if sha1 == lineArr[1] {
		result = true
		return
	}
	err = fmt.Errorf("sha1 check fail expect sha1 %s  but get %s", sha1, lineArr[1])
	return
}
