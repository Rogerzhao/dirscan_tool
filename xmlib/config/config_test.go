package config

import (
	"fmt"
	"testing"
)

var filename = "miui.conf"

func TestConfig(t *testing.T) {
	cfg, err := NewConfig(filename)
	if err != nil {
		t.Fatal(err)
	}

	var database Section
	database, err = cfg.GetSection("database")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(database)
	cfg.DumpConf()
	//fmt.Println(cfg)
}

func TestConfigLine(t *testing.T) {
	cfg, err := NewConfiger("test.conf")
	if err != nil {
		t.Fatal(err)
	}
	cfg.DumpConf()
}
