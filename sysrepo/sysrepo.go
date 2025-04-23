// Package sysrepo provides a Go wrapper for libsysrepo, a YANG-based configuration and operational state data store.
package sysrepo

// #cgo LDFLAGS: -lsysrepo
// #include <stdlib.h>
// #include <sysrepo.h>
// #include <sysrepo/netconf_acm.h>
// #include <sysrepo/subscribed_notifications.h>
import "C"
import (
	"fmt"
	"unsafe"
)

type ErrorCode int

const (
	ErrOk               ErrorCode = C.SR_ERR_OK
	ErrInvalArg         ErrorCode = C.SR_ERR_INVAL_ARG
	ErrLibyang          ErrorCode = C.SR_ERR_LY
	ErrSyscallFailed    ErrorCode = C.SR_ERR_SYS
	ErrNoMemory         ErrorCode = C.SR_ERR_NO_MEMORY
	ErrNotFound         ErrorCode = C.SR_ERR_NOT_FOUND
	ErrExists           ErrorCode = C.SR_ERR_EXISTS
	ErrInternal         ErrorCode = C.SR_ERR_INTERNAL
	ErrUnsupported      ErrorCode = C.SR_ERR_UNSUPPORTED
	ErrValidationFailed ErrorCode = C.SR_ERR_VALIDATION_FAILED
	ErrOperationFailed  ErrorCode = C.SR_ERR_OPERATION_FAILED
	ErrUnauthorized     ErrorCode = C.SR_ERR_UNAUTHORIZED
	ErrLocked           ErrorCode = C.SR_ERR_LOCKED
	ErrTimeout          ErrorCode = C.SR_ERR_TIME_OUT
	ErrCallbackFailed   ErrorCode = C.SR_ERR_CALLBACK_FAILED
	ErrCallbackShelve   ErrorCode = C.SR_ERR_CALLBACK_SHELVE
)

type Error struct {
	Message string
	Code    ErrorCode
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (code: %d)", e.Message, e.Code)
}

func throwIfError(rc C.int, msg string) error {
	if rc != C.SR_ERR_OK {
		return Error{
			Message: msg,
			Code:    ErrorCode(rc),
		}
	}
	return nil
}

// stringToC converts a Go string to a C string and returns a function to free it
func stringToC(s string) (*C.char, func()) {
	if s == "" {
		return nil, func() {}
	}
	cstr := C.CString(s)
	return cstr, func() { C.free(unsafe.Pointer(cstr)) }
}

type Datastore int

const (
	DSRunning        Datastore = C.SR_DS_RUNNING
	DSCandidate      Datastore = C.SR_DS_CANDIDATE
	DSStartup        Datastore = C.SR_DS_STARTUP
	DSOperational    Datastore = C.SR_DS_OPERATIONAL
	DSFactoryDefault Datastore = C.SR_DS_FACTORY_DEFAULT
)

type Event int

const (
	EvChange  Event = C.SR_EV_CHANGE
	EvDone    Event = C.SR_EV_DONE
	EvAbort   Event = C.SR_EV_ABORT
	EvEnabled Event = C.SR_EV_ENABLED
	EvRPC     Event = C.SR_EV_RPC
	EvUpdate  Event = C.SR_EV_UPDATE
)

type NotificationType int

const (
	NotifRealtime       NotificationType = C.SR_EV_NOTIF_REALTIME
	NotifReplay         NotificationType = C.SR_EV_NOTIF_REPLAY
	NotifReplayComplete NotificationType = C.SR_EV_NOTIF_REPLAY_COMPLETE
	NotifTerminated     NotificationType = C.SR_EV_NOTIF_TERMINATED
	NotifModified       NotificationType = C.SR_EV_NOTIF_MODIFIED
	NotifSuspended      NotificationType = C.SR_EV_NOTIF_SUSPENDED
	NotifResumed        NotificationType = C.SR_EV_NOTIF_RESUMED
)

type ChangeOperation int

const (
	OpCreated  ChangeOperation = C.SR_OP_CREATED
	OpModified ChangeOperation = C.SR_OP_MODIFIED
	OpDeleted  ChangeOperation = C.SR_OP_DELETED
	OpMoved    ChangeOperation = C.SR_OP_MOVED
)

type MovePosition int

const (
	MoveBefore MovePosition = C.SR_MOVE_BEFORE
	MoveAfter  MovePosition = C.SR_MOVE_AFTER
	MoveFirst  MovePosition = C.SR_MOVE_FIRST
	MoveLast   MovePosition = C.SR_MOVE_LAST
)

type DefaultOperation string

const (
	OpMerge   DefaultOperation = "merge"
	OpReplace DefaultOperation = "replace"
	OpNone    DefaultOperation = "none"
)

type SubscribeOptions int

const (
	SubsDefault       SubscribeOptions = C.SR_SUBSCR_DEFAULT
	SubsNoThread      SubscribeOptions = C.SR_SUBSCR_NO_THREAD
	SubsPassive       SubscribeOptions = C.SR_SUBSCR_PASSIVE
	SubsDoneOnly      SubscribeOptions = C.SR_SUBSCR_DONE_ONLY
	SubsEnabled       SubscribeOptions = C.SR_SUBSCR_ENABLED
	SubsUpdate        SubscribeOptions = C.SR_SUBSCR_UPDATE
	SubsOperMerge     SubscribeOptions = C.SR_SUBSCR_OPER_MERGE
	SubsThreadSuspend SubscribeOptions = C.SR_SUBSCR_THREAD_SUSPEND
)

type EditOptions int

const (
	EditDefault      EditOptions = C.SR_EDIT_DEFAULT
	EditNonRecursive EditOptions = C.SR_EDIT_NON_RECURSIVE
	EditStrict       EditOptions = C.SR_EDIT_STRICT
	EditIsolate      EditOptions = C.SR_EDIT_ISOLATE
)

type GetOptions int

const (
	GetDefault                       GetOptions = C.SR_OPER_DEFAULT
	GetOperNoState                   GetOptions = C.SR_OPER_NO_STATE
	GetOperNoConfig                  GetOptions = C.SR_OPER_NO_CONFIG
	GetOperNoPullSubscriptions       GetOptions = C.SR_OPER_NO_SUBS
	GetOperNoPushedData              GetOptions = C.SR_OPER_NO_STORED
	GetOperWithOrigin                GetOptions = C.SR_OPER_WITH_ORIGIN
	GetOperNoPollSubscriptionsCached GetOptions = C.SR_OPER_NO_POLL_CACHED
	GetOperNoRunningCached           GetOptions = C.SR_OPER_NO_RUN_CACHED
	GetNoFilter                      GetOptions = C.SR_GET_NO_FILTER
)
