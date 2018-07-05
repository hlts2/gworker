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

	for i := 0; i < 100; i++ {
		d.Add(func() error {
			time.Sleep(time.Second * 1)
			return nil
		})
	}

	d.Start()

	for i := 0; i < 100; i++ {
		d.Add(func() error {
			time.Sleep(time.Second * 2)
			return errors.New("errors.New")
		})
	}

	d.UpScale(20)

FINIS_ALL_JOB:
	for {
		select {
		case err := <-d.JobError():
			fmt.Println(err)
		case _ = <-d.Finish():
			break FINIS_ALL_JOB
		}
	}

	d.Stop()
}
