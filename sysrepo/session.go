package sysrepo

/*
 #cgo LDFLAGS: -lsysrepo
 #include <stdlib.h>
 #include <sysrepo.h>
 #include <sysrepo/netconf_acm.h>

   // Helper function to get value as string for any type
   char* sr_val_to_str(sr_val_t *val) {
    if (val == NULL) {
        return NULL;
    }

    char *mem = NULL;
    size_t size = 0;
    FILE *stream = open_memstream(&mem, &size);
    if (stream == NULL) {
        return NULL;
    }

    switch (val->type) {
        case SR_BINARY_T:
           // Implement
           break;
        case SR_BITS_T:
           // Implement
           break;
        case SR_BOOL_T:
            fprintf(stream, "%s", val->data.bool_val ? "true" : "false");
            break;
        case SR_DECIMAL64_T:
            fprintf(stream, "%f", val->data.decimal64_val);
            break;
        case SR_ENUM_T:
           // Implement
           break;
        case SR_IDENTITYREF_T:
           // Implement
           break;
        case SR_INSTANCEID_T:
           // Implement
           break;
        case SR_INT8_T:
           // Implement
           break;
        case SR_INT16_T:
           // Implement
           break;
        case SR_INT32_T:
           // Implement
           break;
        case SR_INT64_T:
           // Implement
           break;
        case SR_STRING_T:
           if (val->data.string_val) {
                fprintf(stream, "%s", val->data.string_val);
            } else {
                fprintf(stream, "(null)");
            }
           break;
        case SR_UINT8_T:
           // Implement
           break;
        case SR_UINT16_T:
           // Implement
           break;
        case SR_UINT32_T:
           // Implement
           break;
        case SR_UINT64_T:
           // Implement
           break;
        case SR_ANYXML_T:
           // Implement
           break;
        case SR_ANYDATA_T:
           // Implement
           break;
    }

    fclose(stream);
    return mem;
}

*/
import "C"
import (
	"errors"
	"github.com/mattiaswal/go-libyang/libyang"
	"time"
	"unsafe"
)

type Session struct {
	sess         *C.sr_session_ctx_t
	conn         *Connection // Keep reference to connection to prevent GC
	cleanupTasks []func()
}

// Close stops the session
func (s *Session) Close() {
	for i := len(s.cleanupTasks) - 1; i >= 0; i-- {
		s.cleanupTasks[i]()
	}

	if s.sess != nil {
		C.sr_session_stop(s.sess)
		s.sess = nil
	}
}

func (s *Session) GetData(xpath string, maxdepth int, timeout int, opts int) (libyang.DataNode, error) {
	pathC, freePath := stringToC(xpath)
	var dnode *C.sr_data_t
	var dnodePtr **C.sr_data_t

	defer freePath()

	dnodePtr = (**C.sr_data_t)(unsafe.Pointer(&dnode))
	rc := C.sr_get_data(s.sess, pathC, C.uint(maxdepth), C.uint(timeout), C.uint(opts), dnodePtr)

	rawPointer := unsafe.Pointer(dnode.tree)
	node := libyang.NewNode(rawPointer)
	//	C.sr_release_data(dnode) FIX THIS

	return node, throwIfError(rc, "Couldn't get "+xpath)
}

func (s *Session) ActiveDatastore() Datastore {
	return Datastore(C.sr_session_get_ds(s.sess))
}

func (s *Session) SwitchDatastore(datastore Datastore) error {
	rc := C.sr_session_switch_ds(s.sess, C.sr_datastore_t(datastore))
	return throwIfError(rc, "Couldn't switch datastore")
}

func (s *Session) SetItem(path string, value *string, opts EditOptions) error {
	pathC, freePath := stringToC(path)
	defer freePath()

	var valueC *C.char
	var freeValue func()
	if value != nil {
		valueC, freeValue = stringToC(*value)
		defer freeValue()
	}

	rc := C.sr_set_item_str(s.sess, pathC, valueC, nil, C.uint(opts))
	if value != nil {
		return throwIfError(rc, "Couldn't set '"+path+"' to '"+*value+"'")
	}
	return throwIfError(rc, "Couldn't set '"+path+"'")
}

func (s *Session) GetItem(path string) (string, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var value *C.sr_val_t
	rc := C.sr_get_item(s.sess, cPath, 0, &value)
	if rc != 0 { // 0 is SR_ERR_OK
		return "", errors.New("failed to get item")
	}
	defer C.sr_free_val(value)

	cStr := C.sr_val_to_str(value)
	if cStr != nil {
		defer C.free(unsafe.Pointer(cStr))
		return C.GoString(cStr), nil
	}
	return "", errors.New("Value of type  + value._type" + " (unable to convert)")
}

func (s *Session) DeleteItem(path string, opts EditOptions) error {
	pathC, free := stringToC(path)
	defer free()

	rc := C.sr_delete_item(s.sess, pathC, C.uint(opts))
	return throwIfError(rc, "Couldn't delete '"+path+"'")
}

func (s *Session) MoveItem(path string, position MovePosition, keysOrValue *string, origin *string, opts EditOptions) error {
	pathC, freePath := stringToC(path)
	defer freePath()

	var keysOrValueC *C.char
	var freeKeysOrValue func()
	if keysOrValue != nil {
		keysOrValueC, freeKeysOrValue = stringToC(*keysOrValue)
		defer freeKeysOrValue()
	}

	var originC *C.char
	var freeOrigin func()
	if origin != nil {
		originC, freeOrigin = stringToC(*origin)
		defer freeOrigin()
	}

	rc := C.sr_move_item(s.sess, pathC, C.sr_move_position_t(position), keysOrValueC, keysOrValueC, originC, C.uint(opts))
	return throwIfError(rc, "Couldn't move '"+path+"'")
}

func (s *Session) DropForeignOperationalContent(xpath *string) error {
	var xpathC *C.char
	var free func()
	if xpath != nil {
		xpathC, free = stringToC(*xpath)
		defer free()
	}

	rc := C.sr_discard_items(s.sess, xpathC)
	if xpath != nil {
		return throwIfError(rc, "Couldn't discard '"+*xpath+"'")
	}
	return throwIfError(rc, "Couldn't discard all nodes")
}

func (s *Session) ApplyChanges(timeout time.Duration) error {
	rc := C.sr_apply_changes(s.sess, C.uint(timeout/time.Millisecond))
	return throwIfError(rc, "Couldn't apply changes")
}

func (s *Session) DiscardChanges(xpath *string) error {
	var xpathC *C.char
	var free func()
	if xpath != nil {
		xpathC, free = stringToC(*xpath)
		defer free()
	}

	rc := C.sr_discard_changes_xpath(s.sess, xpathC)
	return throwIfError(rc, "Couldn't discard changes")
}

func (s *Session) CopyConfig(source Datastore, moduleName *string, timeout time.Duration) error {
	var moduleNameC *C.char
	var free func()
	if moduleName != nil {
		moduleNameC, free = stringToC(*moduleName)
		defer free()
	}

	rc := C.sr_copy_config(s.sess, moduleNameC, C.sr_datastore_t(source), C.uint(timeout/time.Millisecond))
	return throwIfError(rc, "Couldn't copy config")
}

type ErrorInfo struct {
	Code    ErrorCode
	Message string
}

func (s *Session) GetOriginatorName() string {
	return C.GoString(C.sr_session_get_orig_name(s.sess))
}

func (s *Session) SetOriginatorName(originatorName string) error {
	nameC, free := stringToC(originatorName)
	defer free()

	rc := C.sr_session_set_orig_name(s.sess, nameC)
	return throwIfError(rc, "Couldn't set originator name")
}

func (s *Session) GetConnection() *Connection {
	return s.conn
}

func (s *Session) GetId() uint32 {
	return uint32(C.sr_session_get_id(s.sess))
}

func (s *Session) SetNacmUser(user string) error {
	userC, free := stringToC(user)
	defer free()

	rc := C.sr_nacm_set_user(s.sess, userC)
	return throwIfError(rc, "Couldn't set NACM user")
}

func (s *Session) GetNacmUser() *string {
	userName := C.sr_nacm_get_user(s.sess)
	if userName == nil {
		return nil
	}
	result := C.GoString(userName)
	return &result
}

func GetNacmRecoveryUser() string {
	return C.GoString(C.sr_nacm_get_recovery_user())
}
