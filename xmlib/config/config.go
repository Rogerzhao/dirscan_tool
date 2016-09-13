package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// Section of setting
type Section map[string]string

// Configer interface definition of configure
type Configer interface {
	GetSection(sectionName string) (section Section, err error)
	GetAllSections() (sectionList map[string]Section)
	GetSetting(sectionName, keyName string) (value string, err error)
	GetBoolSetting(sectionName, keyName string, dfault bool) (value bool, err error)
	GetIntSetting(sectionName, keyName string, dfault int64) (value int64, err error)
	LastModify() time.Time
	ConfigureFile() string
	DumpConf() (err error)
}

// Config struct holding all configure information
type Config struct {
	sections   map[string]Section
	lastModify time.Time
	fileName   string
}

// NewConfigFromReader create Config from reader(file/stdin/zk/ etc...)
func NewConfigFromReader(r io.Reader) (cfgr Configer, err error) {
	inputReader := bufio.NewReader(r)
	cfg, err := readConfig(inputReader)
	if err != nil {
		return
	}
	cfgr = cfg
	return
}

// NewConfigFromMap create Config from etcdMap
func NewConfigFromEtcdMap(prefix string, etcdMap map[string]string) (cfg *Config, err error) {
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	cfg = new(Config)
	cfg.sections = make(map[string]Section)
	for key, value := range etcdMap {
		key = strings.Replace(key, prefix, "", 1)
		paths := strings.Split(key, "/")
		if len(paths) != 2 {
			err = fmt.Errorf("key error : %s .", key)
			return
		}
		curSection := paths[0]
		if _, ok := cfg.sections[curSection]; !ok {
			cfg.sections[curSection] = make(Section)
		}
		cfg.sections[curSection][paths[1]] = value
	}
	return
}

// NewConfiger create Configer from filename
func NewConfiger(fileName string) (cfg Configer, err error) {
	return NewConfig(fileName)
}

// NewConfig create Config from filename
func NewConfig(fileName string) (cfg *Config, err error) {
	var file *os.File
	file, err = os.Open(fileName)
	if err != nil {
		return
	}
	defer file.Close()
	inputReader := bufio.NewReader(file)
	cfg, err = readConfig(inputReader)

	var st os.FileInfo
	st, err = file.Stat()
	if err != nil {
		return
	}
	cfg.lastModify = st.ModTime()
	cfg.fileName = fileName
	return
}

func readConfig(reader *bufio.Reader) (cfg *Config, err error) {
	curSection := ""
	cfg = new(Config)
	cfg.sections = make(map[string]Section)
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err == io.EOF && line != "" {
			err = nil
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}
		line = strings.Trim(line, "\r\n ")
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		switch {
		case line[0] == ';' || line[0] == '#': // comment
			continue
		case line[0] == '[': //section
			curSection = strings.Trim(line, "[] ")
		default: // A = B
			var key, value string
			if strings.Contains(line, "=") {
				pair := strings.Index(line, "=")
				// key = strings.TrimSpace(pair[0])
				// value = strings.TrimSpace(pair[1])
				key = strings.TrimSpace(line[:pair])
				value = strings.TrimSpace(strings.Replace(line[pair:], "=", "", 1))
			} else {
				key = strings.TrimSpace(line)
				value = ""
			}
			if _, ok := cfg.sections[curSection]; !ok {
				cfg.sections[curSection] = make(Section)
			}
			cfg.sections[curSection][key] = value
		}
	}
	err = nil
	return
}

// GetSection get a section by section name
func (c *Config) GetSection(sectionName string) (section Section, err error) {
	var ok bool
	var sectionReal Section
	section = make(Section)
	if sectionReal, ok = c.sections[sectionName]; !ok {
		err = fmt.Errorf("Section %s not exists.", sectionName)
		return
	}
	for k, v := range sectionReal {
		section[k] = v
	}
	return
}

// GetAllSections get all section of Config
func (c *Config) GetAllSections() (sectionList map[string]Section) {
	sectionList = make(map[string]Section)
	for sectionName, section := range c.sections {
		sectionList[sectionName] = section
	}
	return
}

// GetSetting get value from speccific section and key
func (c *Config) GetSetting(sectionName, keyName string) (value string, err error) {
	var ok bool
	if _, ok = c.sections[sectionName]; !ok {
		err = fmt.Errorf("Section %s not exists.", sectionName)
		return
	}
	if _, ok = c.sections[sectionName][keyName]; !ok {
		err = fmt.Errorf("Section %s key %s not exists.", sectionName, keyName)
		return
	}
	err = nil
	value = c.sections[sectionName][keyName]
	return
}

// GetIntSetting get setting and convert to int64
func (c *Config) GetIntSetting(sectionName, keyName string, defaultVal int64) (value int64, err error) {
	var s string
	if s, err = c.GetSetting(sectionName, keyName); err != nil {
		return defaultVal, err
	}
	if value, err = strconv.ParseInt(s, 10, 64); err != nil {
		return defaultVal, err
	}
	return value, nil
}

// GetBoolSetting get setting and convert to bool
func (c *Config) GetBoolSetting(sectionName, keyName string, dfault bool) (value bool, err error) {
	var s string
	if s, err = c.GetSetting(sectionName, keyName); err != nil {
		return dfault, err
	}
	if value, err = strconv.ParseBool(s); err != nil {

		return dfault, err
	}
	if value, err = strconv.ParseBool(s); err != nil {
		return dfault, err
	}
	return value, nil
}

// LastModify return lastmodify time
func (c *Config) LastModify() time.Time {
	return c.lastModify
}

// ConfigureFile return backend file
func (c *Config) ConfigureFile() string {
	return c.fileName
}

// GetValue of section, return string
func (c *Section) GetValue(keyName string) (value string, err error) {
	var ok bool
	if _, ok = (*c)[keyName]; !ok {
		err = fmt.Errorf("Key %s not exists.", keyName)
	}
	value = (*c)[keyName]
	return
}

// GetIntValue of section, convert to int
func (c *Section) GetIntValue(keyName string) (value int, err error) {
	valueStr, err := c.GetValue(keyName)
	if err != nil {
		return
	}
	value, err = strconv.Atoi(valueStr)
	return
}

// GetInt64Value get value of section, convert to int64
func (c *Section) GetInt64Value(keyName string) (value int64, err error) {
	valueStr, err := c.GetValue(keyName)
	if err != nil {
		return
	}
	value, err = strconv.ParseInt(valueStr, 10, 64)
	return
}

// GetFloatValue get value of section, convert to int64
func (c *Section) GetFloatValue(keyName string) (value float64, err error) {
	valueStr, err := c.GetValue(keyName)
	if err != nil {
		return
	}
	value, err = strconv.ParseFloat(valueStr, 64)
	return
}

// GetBoolValue get value of section, covnert to bool
func (c *Section) GetBoolValue(keyName string) (value bool, err error) {
	valueStr, err := c.GetValue(keyName)
	if err != nil {
		return
	}
	value, err = strconv.ParseBool(valueStr)
	return
}

// DumpConf dump configure of Config
func (c *Config) DumpConf() (err error) {
	for sectionName, section := range c.sections {
		fmt.Printf("[%s]\n", sectionName)
		for key, value := range section {
			fmt.Printf("%s=%s\n", key, value)
		}
		fmt.Println()
	}
	return
}
