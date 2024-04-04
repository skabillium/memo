package db

type MemoObjType = byte

const (
	ObjValue MemoObjType = iota
	ObjQueue
	ObjPQueue
	ObjList
)

type MemoObj struct {
	Kind   MemoObjType
	Value  string
	Queue  *Queue
	PQueue *PriorityQueue
	List   *List
}

func (d *Database) newValueObj(value string) *MemoObj {
	return &MemoObj{Kind: ObjValue, Value: value}
}

func (d *Database) newQueueObj() *MemoObj {
	return &MemoObj{Kind: ObjQueue, Queue: NewQueue()}
}

func (d *Database) newPQueueObj() *MemoObj {
	return &MemoObj{Kind: ObjPQueue, PQueue: NewPriorityQueue()}
}

func (d *Database) newListObj() *MemoObj {
	return &MemoObj{Kind: ObjList, List: NewList()}
}

func (obj *MemoObj) asValue() (string, bool) {
	return obj.Value, obj.Kind == ObjValue
}

func (obj *MemoObj) asQueue() (*Queue, bool) {
	return obj.Queue, obj.Kind == ObjQueue
}

func (obj *MemoObj) asPQueue() (*PriorityQueue, bool) {
	return obj.PQueue, obj.Kind == ObjPQueue
}

func (obj *MemoObj) asList() (*List, bool) {
	return obj.List, obj.Kind == ObjList
}
