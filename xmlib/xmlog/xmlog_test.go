package xmlog

import (
	"log"
	"testing"
	"time"
)

func TestLogrotate(t *testing.T) {
	log.Println("start test log rotate")
	Watch("xmlogtest", ".")
	for i := 0; ; i++ {
		now := time.Now()
		Errorf("i = %v", now)
		time.Sleep(1 * time.Second)
	}
	Close()
	log.Println("end test log rotate")
}
