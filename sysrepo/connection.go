package sysrepo

// #cgo LDFLAGS: -lsysrepo
// #include <stdlib.h>
// #include <sysrepo.h>
import "C"
import (
	"runtime"
	"time"
)

type Connection struct {
	conn *C.sr_conn_ctx_t
}

type ConnectionFlag int

const (
	ConnDefault           ConnectionFlag = C.SR_CONN_DEFAULT
	ConnCacheRunning      ConnectionFlag = C.SR_CONN_CACHE_RUNNING
	ConnLibYangPrivParsed ConnectionFlag = C.SR_CONN_CTX_SET_PRIV_PARSED
)

// Connect creates a new connection to the sysrepo datastore
func Connect(options ConnectionFlag) (*Connection, error) {
	var conn *C.sr_conn_ctx_t
	rc := C.sr_connect(C.sr_conn_options_t(options), &conn)
	if rc != C.SR_ERR_OK {
		return nil, Error{
			Message: "Couldn't connect to sysrepo",
			Code:    ErrorCode(rc),
		}
	}

	connection := &Connection{conn: conn}
	runtime.SetFinalizer(connection, (*Connection).Close)
	return connection, nil
}

func (c *Connection) Close() {
	if c.conn != nil {
		C.sr_disconnect(c.conn)
		c.conn = nil
	}
}

func (c *Connection) SessionStart(datastore Datastore) (*Session, error) {
	var sess *C.sr_session_ctx_t
	rc := C.sr_session_start(c.conn, C.sr_datastore_t(datastore), &sess)
	if rc != C.SR_ERR_OK {
		return nil, Error{
			Message: "Couldn't start sysrepo session",
			Code:    ErrorCode(rc),
		}
	}

	session := &Session{
		sess: sess,
		conn: c,
	}
	runtime.SetFinalizer(session, (*Session).Close)
	return session, nil
}

func (c *Connection) SetModuleReplaySupport(moduleName string, enabled bool) error {
	moduleNameC, free := stringToC(moduleName)
	defer free()
	var cEnabled C.int
	if enabled {
		cEnabled = 1
	} else {
		cEnabled = 0
	}

	rc := C.sr_set_module_replay_support(c.conn, moduleNameC, cEnabled)
	return throwIfError(rc, "Couldn't set replay support for module '"+moduleName+"'")
}

type ModuleReplaySupport struct {
	Enabled       bool
	EarliestNotif *time.Time
}

func (c *Connection) GetModuleReplaySupport(moduleName string) (*ModuleReplaySupport, error) {
	moduleNameC, free := stringToC(moduleName)
	defer free()

	var enabled C.int
	var earliestNotif C.struct_timespec

	rc := C.sr_get_module_replay_support(c.conn, moduleNameC, &earliestNotif, &enabled)
	if rc != C.SR_ERR_OK {
		return nil, Error{
			Message: "Couldn't get replay support for module '" + moduleName + "'",
			Code:    ErrorCode(rc),
		}
	}

	result := &ModuleReplaySupport{
		Enabled: enabled != 0,
	}

	if earliestNotif.tv_sec != 0 || earliestNotif.tv_nsec != 0 {
		t := time.Unix(int64(earliestNotif.tv_sec), int64(earliestNotif.tv_nsec))
		result.EarliestNotif = &t
	}

	return result, nil
}

func (c *Connection) DiscardOperationalChanges(xpath string, session *Session, timeout time.Duration) error {
	xpathC, freeXpath := stringToC(xpath)
	defer freeXpath()

	var sessPtr *C.sr_session_ctx_t
	if session != nil {
		sessPtr = session.sess
	}

	rc := C.sr_discard_oper_changes(c.conn, sessPtr, xpathC, C.uint(timeout/time.Millisecond))
	return throwIfError(rc, "Couldn't discard operational changes")
}
