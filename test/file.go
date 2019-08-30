package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/zkynetio/safelocker"
)

type LockableFile struct {
	File *os.File
	safelocker.SafeLocker
}

func doshit() {

	go func() {
		lf := LockableFile{}
		var err error
		lf.File, err = os.OpenFile("text.log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			log.Println(err)
		}

		defer lf.File.Close()
		for i := 0; i < 100000; i++ {
			lf.Lock()
			time.Sleep(1 * time.Second)
			if _, err := lf.File.WriteString("text to append" + strconv.Itoa(i) + "\n"); err != nil {
				log.Println(err)
			}
			lf.Unlock()
		}

	}()

	go func() {
		follow("text.log")
	}()

	for {
		time.Sleep(2 * time.Second)
	}
}

func follow(filename string) error {
	file, _ := os.Open(filename)
	fd, _ := syscall.InotifyInit()
	_, _ = syscall.InotifyAddWatch(fd, filename, syscall.IN_MODIFY)
	r := bufio.NewReader(file)
	for {
		by, err := r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		fmt.Print(string(by))
		if err != io.EOF {
			continue
		}
		if err = waitForChange(fd); err != nil {
			return err
		}
	}
}

func waitForChange(fd int) error {
	for {
		var buf [syscall.SizeofInotifyEvent]byte
		_, err := syscall.Read(fd, buf[:])
		if err != nil {
			return err
		}
		r := bytes.NewReader(buf[:])
		var ev = syscall.InotifyEvent{}
		_ = binary.Read(r, binary.LittleEndian, &ev)
		if ev.Mask&syscall.IN_MODIFY == syscall.IN_MODIFY {
			return nil
		}
	}
}
