package util

import (
	"github.com/sirupsen/logrus"
)

// SafeGo runs fn in a new goroutine and logs recovered panics. Use for post-response
// background work (email, notifications) so a single bug cannot crash the process.
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("recovered", r).Error("background goroutine panicked")
			}
		}()
		fn()
	}()
}
