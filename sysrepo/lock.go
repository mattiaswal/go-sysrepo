package sysrepo

// #cgo LDFLAGS: -lsysrepo
// #include <stdlib.h>
// #include <sysrepo.h>
import "C"
import (
	"runtime"
	"time"
)

type Lock struct {
	session  *Session
	lockedDs Datastore
	module   *string
}

func NewLock(session *Session, moduleName *string, timeout *time.Duration) (*Lock, error) {
	var moduleNameC *C.char
	var free func()
	if moduleName != nil {
		moduleNameC, free = stringToC(*moduleName)
		defer free()
	}

	var timeoutC C.uint
	if timeout != nil {
		timeoutC = C.uint(timeout.Milliseconds())
	}

	rc := C.sr_lock(session.sess, moduleNameC, timeoutC)
	if rc != C.SR_ERR_OK {
		return nil, Error{
			Message: "Cannot lock session",
			Code:    ErrorCode(rc),
		}
	}

	lock := &Lock{
		session:  session,
		lockedDs: session.ActiveDatastore(),
		module:   moduleName,
	}

	runtime.SetFinalizer(lock, (*Lock).Unlock)
	return lock, nil
}

func (l *Lock) Unlock() error {
	if l.session == nil {
		return nil // Already unlocked
	}

	currentDs := l.session.ActiveDatastore()

	err := l.session.SwitchDatastore(l.lockedDs)
	if err != nil {
		return err
	}

	var moduleNameC *C.char
	var free func()
	if l.module != nil {
		moduleNameC, free = stringToC(*l.module)
		defer free()
	}

	rc := C.sr_unlock(l.session.sess, moduleNameC)

	l.session.SwitchDatastore(currentDs)

	if rc != C.SR_ERR_OK {
		return Error{
			Message: "Cannot unlock session",
			Code:    ErrorCode(rc),
		}
	}

	l.session = nil // Mark as unlocked
	return nil
}
