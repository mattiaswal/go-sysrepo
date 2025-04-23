package sysrepo

// #cgo LDFLAGS: -lsysrepo
// #include <stdlib.h>
// #include <sysrepo.h>
import "C"
import (
	"runtime"
)

type Change struct {
	Operation       ChangeOperation
	PreviousValue   *string
	PreviousList    *string
	PreviousDefault bool
}

type ChangeCollection struct {
	xpath string
	sess  *Session
}

type ChangeIterator struct {
	iter    *C.sr_change_iter_t
	session *Session
	current *Change
}

func NewChangeCollection(session *Session, xpath string) *ChangeCollection {
	return &ChangeCollection{
		xpath: xpath,
		sess:  session,
	}
}

func (c *ChangeCollection) Begin() (*ChangeIterator, error) {
	xpathC, freeXpath := stringToC(c.xpath)
	defer freeXpath()

	var iter *C.sr_change_iter_t
	rc := C.sr_get_changes_iter(c.sess.sess, xpathC, &iter)
	if rc != C.SR_ERR_OK {
		return nil, Error{
			Message: "Couldn't create an iterator for changes",
			Code:    ErrorCode(rc),
		}
	}

	iterator := &ChangeIterator{
		iter:    iter,
		session: c.sess,
	}

	runtime.SetFinalizer(iterator, (*ChangeIterator).Close)

	err := iterator.Next()
	if err != nil {
		iterator.Close()
		return nil, err
	}

	return iterator, nil
}

func (i *ChangeIterator) Close() {
	if i.iter != nil {
		C.sr_free_change_iter(i.iter)
		i.iter = nil
	}
}

func (i *ChangeIterator) Next() error {
	if i.iter == nil {
		i.current = nil
		return nil
	}

	var operation C.sr_change_oper_t
	var node *C.struct_lyd_node
	var prevValue, prevList *C.char
	var prevDefault C.int

	rc := C.sr_get_change_tree_next(
		i.session.sess,
		i.iter,
		&operation,
		&node,
		&prevValue,
		&prevList,
		&prevDefault)

	if rc == C.SR_ERR_NOT_FOUND {
		i.current = nil
		return nil
	}

	if rc != C.SR_ERR_OK {
		return Error{
			Message: "Could not iterate to the next change",
			Code:    ErrorCode(rc),
		}
	}

	// Convert previous value
	var goPrevValue *string
	if prevValue != nil {
		s := C.GoString(prevValue)
		goPrevValue = &s
	}

	// Convert previous list
	var goPrevList *string
	if prevList != nil {
		s := C.GoString(prevList)
		goPrevList = &s
	}

	i.current = &Change{
		Operation:       ChangeOperation(operation),
		PreviousValue:   goPrevValue,
		PreviousList:    goPrevList,
		PreviousDefault: prevDefault != 0,
	}

	return nil
}

func (i *ChangeIterator) HasNext() bool {
	return i.current != nil
}

func (i *ChangeIterator) Current() *Change {
	return i.current
}

func (s *Session) GetChanges(xpath string) *ChangeCollection {
	return NewChangeCollection(s, xpath)
}
