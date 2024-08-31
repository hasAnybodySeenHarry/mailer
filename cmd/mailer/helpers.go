package main

import (
	"fmt"
	"time"
)

func (app *application) retry(times int, wait time.Duration, fn func() error) error {
	var err error

	for i := 0; i < times; i++ {
		err = fn()
		if nil == err {
			return nil
		}

		time.Sleep(wait)
	}

	return err
}

func (app *application) background(fn func()) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.Println(fmt.Errorf("%s", err))
			}
		}()

		fn()
	}()
}
