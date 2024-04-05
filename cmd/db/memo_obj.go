package db

import "time"

type MemoObjType = byte

const (
	ObjValue MemoObjType = iota
	ObjPQueue
	ObjList
)

type MemoObj struct {
	Kind      MemoObjType
	Value     string
	ExpiresAt int64
	PQueue    *PriorityQueue
	List      *List
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

func (obj *MemoObj) asValue() (string, bool) {
	return obj.Value, obj.Kind == ObjValue
}

func (obj *MemoObj) asPQueue() (*PriorityQueue, bool) {
	return obj.PQueue, obj.Kind == ObjPQueue
}

func (obj *MemoObj) asList() (*List, bool) {
	return obj.List, obj.Kind == ObjList
}

// Set expiration for object in seconds
func (obj *MemoObj) ExpireIn(seconds int) {
	obj.ExpiresAt = time.Now().Unix() + int64(seconds)
}

func (obj *MemoObj) hasExpired() bool {
	return obj.ExpiresAt != 0 && obj.ExpiresAt < time.Now().Unix()
}
