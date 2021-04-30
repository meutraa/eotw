package input

// #include <linux/input-event-codes.h>
// #include <linux/input.h>
import "C"

import (
	"encoding/binary"
	"log"
	"os"
	"syscall"
)

type keyEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

type Event struct {
	Pressed  bool
	Released bool
	//https://github.com/torvalds/linux/blob/master/include/uapi/linux/input-event-codes.h
	Code uint16
	Time syscall.Timeval
}

func ReadInput(kbd string, events chan *Event) error {
	file, err := os.Open(kbd)
	if err != nil {
		return err
	}
	go func() {
		defer file.Close()

		var ev keyEvent
		for {
			err = binary.Read(file, binary.LittleEndian, &ev)
			if nil != err {
				log.Println(err, "unable to read keyboard input")
				return
			}
			if ev.Type != C.EV_KEY {
				continue
			}
			events <- &Event{
				Pressed:  ev.Value == 1,
				Released: ev.Value == 0,
				Code:     ev.Code,
				Time:     ev.Time,
			}
		}
	}()
	return nil
}
