package db

import "errors"

var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

type Database struct {
	objs map[string]*MemoObj
}

func NewDatabase() *Database {
	return &Database{objs: make(map[string]*MemoObj)}
}

func (d *Database) FlushAll() {
	d.objs = make(map[string]*MemoObj)
}

func (d *Database) Keys() []string {
	keys := make([]string, len(d.objs))
	for k := range d.objs {
		keys = append(keys, k)
	}

	return keys
}

func (d *Database) Get(key string) (string, bool, error) {
	obj, found := d.objs[key]
	if !found {
		return "", found, nil
	}

	value, ok := obj.asValue()
	if !ok {
		return "", found, ErrWrongType
	}

	return value, found, nil
}

func (d *Database) Set(key string, value string) {
	d.objs[key] = d.newValueObj(value)
}

func (d *Database) Del(key string) {
	delete(d.objs, key)
}

func (d *Database) Qadd(qname string, value string) error {
	obj, found := d.objs[qname]
	if !found {
		obj = d.newQueueObj()
	}

	queue, ok := obj.asQueue()
	if !ok {
		return ErrWrongType
	}

	queue.Enqueue(value)
	if !found {
		d.objs[qname] = obj
	}

	return nil
}

func (d *Database) QPop(qname string) (string, bool, error) {
	obj, found := d.objs[qname]
	if !found {
		return "", false, nil
	}

	queue, ok := obj.asQueue()
	if !ok {
		return "", found, ErrWrongType
	}

	value := queue.Dequeue()
	if queue.Length == 0 {
		d.Del(qname)
	}

	return value, true, nil
}

func (d *Database) Qlen(qname string) (int, bool, error) {
	obj, found := d.objs[qname]
	if !found {
		return -1, found, nil
	}

	queue, ok := obj.asQueue()
	if !ok {
		return -1, found, ErrWrongType
	}

	return queue.Length, found, nil
}

func (d *Database) PQAdd(qname string, value string, priority int) error {
	obj, found := d.objs[qname]
	if !found {
		obj = d.newPQueueObj()
	}

	pqueue, ok := obj.asPQueue()
	if !ok {
		return ErrWrongType
	}

	pqueue.Enqueue(value, priority)
	if !found {
		d.objs[qname] = obj
	}

	return nil
}

func (d *Database) PQPop(qname string) (string, bool, error) {
	obj, found := d.objs[qname]
	if !found {
		return "", found, nil
	}

	pqueue, ok := obj.asPQueue()
	if !ok {
		return "", found, ErrWrongType
	}

	value := pqueue.Dequeue()
	if pqueue.Length == 0 {
		d.Del(qname)
	}

	return value, found, nil
}

func (d *Database) PQLen(qname string) (int, bool, error) {
	obj, found := d.objs[qname]
	if !found {
		return -1, found, nil
	}

	pqueue, ok := obj.asPQueue()
	if !ok {
		return -1, found, ErrWrongType
	}

	return pqueue.Length, found, nil
}

func (d *Database) LPush(lname string, value string) error {
	obj, found := d.objs[lname]
	if !found {
		obj = d.newListObj()
	}

	list, ok := obj.asList()
	if !ok {
		return ErrWrongType
	}

	list.Prepend(value)
	if !found {
		d.objs[lname] = obj
	}

	return nil
}

func (d *Database) RPush(lname string, value string) error {
	obj, found := d.objs[lname]
	if !found {
		obj = d.newListObj()
	}

	list, ok := obj.asList()
	if !ok {
		return ErrWrongType
	}

	list.Append(value)
	if !found {
		d.objs[lname] = obj
	}

	return nil
}

func (d *Database) LPop(lname string) (string, bool, error) {
	obj, found := d.objs[lname]
	if !found {
		return "", found, nil
	}

	list, ok := obj.asList()
	if !ok {
		return "", found, ErrWrongType
	}

	value := list.PopHead()
	if list.Length == 0 {
		d.Del(lname)
	}

	return value, found, nil
}

func (d *Database) RPop(lname string) (string, bool, error) {
	obj, found := d.objs[lname]
	if !found {
		return "", found, nil
	}

	list, ok := obj.asList()
	if !ok {
		return "", found, ErrWrongType
	}

	value := list.PopTail()
	if obj.List.Length == 0 {
		d.Del(lname)
	}

	return value, found, nil
}

func (d *Database) LLen(lname string) (int, error) {
	obj, found := d.objs[lname]
	if !found {
		return -1, nil
	}

	list, ok := obj.asList()
	if !ok {
		return -1, ErrWrongType
	}

	return list.Length, nil
}
