package db

import (
	"errors"
	"path/filepath"
)

var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

type Database struct {
	objs map[string]*MemoObj
}

func NewDatabase() *Database {
	return &Database{objs: make(map[string]*MemoObj)}
}

func (d *Database) Size() int {
	return len(d.objs)
}

func (d *Database) FlushAll() {
	d.objs = make(map[string]*MemoObj)
}

func (d *Database) CleanupExpired() int {
	deleted := 0
	for k, obj := range d.objs {
		if obj.hasExpired() {
			delete(d.objs, k)
			deleted++
		}
	}

	return deleted
}

func (d *Database) Keys(pattern string) []string {
	keys := []string{}
	for k := range d.objs {
		match, err := filepath.Match(pattern, k)
		if err != nil {
			break
		}

		if match {
			keys = append(keys, k)
		}
	}

	return keys
}

func (d *Database) Expire(key string, seconds int) bool {
	obj, found := d.getObj(key)
	if !found {
		return found
	}
	if seconds <= 0 {
		d.Del(key)
		return true
	}

	obj.ExpireIn(seconds)
	return true
}

func (d *Database) Get(key string) (string, bool, error) {
	obj, found := d.getObj(key)
	if !found {
		return "", found, nil
	}

	value, ok := obj.asValue()
	if !ok {
		return "", found, ErrWrongType
	}

	return value, found, nil
}

func (d *Database) Set(key string, value string, expires int) {
	obj := newValueObj(value)
	obj.ExpireIn(expires)
	d.objs[key] = obj
}

func (d *Database) Del(key string) {
	delete(d.objs, key)
}

func (d *Database) PQAdd(qname string, value string, priority int) error {
	obj, found := d.getObj(qname)
	if !found {
		obj = newPQueueObj()
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
	obj, found := d.getObj(qname)
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
	obj, found := d.getObj(qname)
	if !found {
		return -1, found, nil
	}

	pqueue, ok := obj.asPQueue()
	if !ok {
		return -1, found, ErrWrongType
	}

	return pqueue.Length, found, nil
}

func (d *Database) LPush(lname string, values []string) error {
	obj, found := d.getObj(lname)
	if !found {
		obj = newListObj()
	}

	list, ok := obj.asList()
	if !ok {
		return ErrWrongType
	}

	for i := 0; i < len(values); i++ {
		list.Prepend(values[i])
	}
	if !found {
		d.objs[lname] = obj
	}

	return nil
}

func (d *Database) RPush(lname string, values []string) error {
	obj, found := d.getObj(lname)
	if !found {
		obj = newListObj()
	}

	list, ok := obj.asList()
	if !ok {
		return ErrWrongType
	}

	for i := 0; i < len(values); i++ {
		list.Append(values[i])
	}
	if !found {
		d.objs[lname] = obj
	}

	return nil
}

func (d *Database) LPop(lname string) (string, bool, error) {
	obj, found := d.getObj(lname)
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
	obj, found := d.getObj(lname)
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
	obj, found := d.getObj(lname)
	if !found {
		return -1, nil
	}

	list, ok := obj.asList()
	if !ok {
		return -1, ErrWrongType
	}

	return list.Length, nil
}

func (d *Database) getObj(key string) (*MemoObj, bool) {
	obj, found := d.objs[key]
	if !found {
		return nil, found
	}

	if obj.hasExpired() {
		d.Del(key)
		return nil, false
	}

	return obj, true
}
