package util

import (
	"log"
	"os"
	"time"
)

func GetNowMsTime() int64 {
	return time.Now().UnixNano() / 1e6
}

func Int64Time2String(t int64) string {
	tm := time.Unix(t/1000, 0)
	return tm.Format("2006-01-02/15:04:05")
}

//type OnlyOneThread struct {
//	l sync.Mutex
//	o bool
//}

//func (oot *OnlyOneThread) Do(f func(p interface{}) (r interface{}), p interface{}) (r interface{}) {
//	if oot.o {
//		return errors.New("thread is running")
//	}
//	oot.l.Lock()
//	defer oot.l.Unlock()
//	if oot.o {
//		return errors.New("thread is running")
//	}
//	oot.o = true
//	defer func() {
//		oot.o = false
//	}()
//
//	return f(p)
//}
//
//func (oot *OnlyOneThread) Do2(f func()) {
//	if oot.o {
//		return
//	}
//	oot.l.Lock()
//	defer oot.l.Unlock()
//	if oot.o {
//		return
//	}
//	oot.o = true
//	defer func() {
//		oot.o = false
//	}()
//
//	f()
//}
//
//func (oot *OnlyOneThread) Start() bool {
//	if oot.o {
//		return false
//	}
//	oot.l.Lock()
//	defer oot.l.Unlock()
//	if oot.o {
//		return false
//	}
//	oot.o = true
//
//	return true
//}
//
//func (oot *OnlyOneThread) Release() {
//	oot.l.Lock()
//	defer oot.l.Unlock()
//
//	oot.o = false
//}

func Save2File(data []byte, filename string) error {

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		f.Close()
		log.Fatal(err)
	}

	return nil
}
