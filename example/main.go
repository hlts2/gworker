package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/hlts2/gworker"
)

func main() {
	d := gworker.NewDispatcher(3)
	d.StartJobObserver()

	for i := 0; i < 50; i++ {
		d.Add(func() error {
			time.Sleep(time.Second * 4)
			return errors.New("errors.New")
		})
	}
	d.Start()

	time.Sleep(time.Second * 10)

	d.UpScale(100)

END_LOOP:
	for {
		select {
		case err := <-d.JobError():
			fmt.Println(err)
		case _ = <-d.Finish():
			break END_LOOP
		}
	}
}
