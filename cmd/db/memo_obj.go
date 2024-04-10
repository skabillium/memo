package db

import "time"

type MemoObjType = byte

const (
	ObjValue MemoObjType = iota
	ObjPQueue
	ObjList
	ObjSet
)

// The base object that is stored in the database, all Memo data structures are a type of
// MemoObj. Some utility functions are also provided for casting the object to specific data
// structures.
type MemoObj struct {
	Kind      MemoObjType
	Value     string
	ExpiresAt int64
	PQueue    *PriorityQueue
	List      *List
	Set       *Set
}

func newValueObj(value string) *MemoObj {
	return &MemoObj{Kind: ObjValue, Value: value}
}

func newPQueueObj() *MemoObj {
	return &MemoObj{Kind: ObjPQueue, PQueue: NewPriorityQueue()}
}

func newListObj() *MemoObj {
	return &MemoObj{Kind: ObjList, List: NewList()}
}

func newSetObj() *MemoObj {
	return &MemoObj{Kind: ObjSet, Set: NewSet()}
}

func (obj *MemoObj) asValue() (string, bool) {
	return obj.Value, obj.Kind == ObjValue
}

func (obj *MemoObj) asPQueue() (*PriorityQueue, bool) {
	return obj.PQueue, obj.Kind == ObjPQueue
}

func (obj *MemoObj) asList() (*List, bool) {
	return obj.List, obj.Kind == ObjList
}

func (obj *MemoObj) asSet() (*Set, bool) {
	return obj.Set, obj.Kind == ObjSet
}

// Set expiration for object in seconds
func (obj *MemoObj) ExpireIn(seconds int) {
	obj.ExpiresAt = time.Now().Unix() + int64(seconds)
}

// Check if object has expired
func (obj *MemoObj) hasExpired() bool {
	return obj.ExpiresAt != 0 && obj.ExpiresAt < time.Now().UnixMilli()
}
