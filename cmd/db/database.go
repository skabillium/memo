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

func (d *Database) CleanupExpired(limit int) int {
	deleted := 0
	for k, obj := range d.objs {
		if obj.hasExpired() {
			delete(d.objs, k)
			deleted++
		}

		if limit != 0 && deleted == limit {
			break
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
	if expires != 0 {
		obj.ExpireIn(expires)
	}
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

func (d *Database) SetAdd(key string, values []string) (int, error) {
	obj, found := d.getObj(key)
	if !found {
		obj = newSetObj()
	}

	set, ok := obj.asSet()
	if !ok {
		return -1, ErrWrongType
	}

	for i := 0; i < len(values); i++ {
		set.Add(values[i])
	}

	if !found {
		d.objs[key] = obj
	}

	return len(values), nil
}

func (d *Database) SetMembers(key string) ([]string, error) {
	obj, found := d.getObj(key)
	if !found {
		obj = newSetObj()
	}

	set, ok := obj.asSet()
	if !ok {
		return nil, ErrWrongType
	}

	return set.Items(), nil
}

func (d *Database) SetRemove(key string, values []string) (int, error) {
	obj, found := d.getObj(key)
	if !found {
		return 0, nil
	}

	set, ok := obj.asSet()
	if !ok {
		return -1, ErrWrongType
	}

	var total int
	for i := 0; i < len(values); i++ {
		removed := set.Delete(values[i])
		if removed {
			total++
		}
	}

	return total, nil
}

func (d *Database) SetIsMember(key string, value string) (bool, error) {
	obj, found := d.getObj(key)
	if !found {
		return false, nil
	}

	set, ok := obj.asSet()
	if !ok {
		return false, ErrWrongType
	}

	return set.Has(value), nil
}

func (d *Database) SetCard(key string) (int, error) {
	obj, found := d.getObj(key)
	if !found {
		return 0, nil
	}

	set, ok := obj.asSet()
	if !ok {
		return -1, ErrWrongType
	}

	return set.Size, nil
}

func (d *Database) SetInter(a string, b string) ([]string, error) {
	first, found := d.getObj(a)
	if !found {
		return nil, nil
	}

	second, found := d.getObj(b)
	if !found {
		return nil, nil
	}

	set1, ok := first.asSet()
	if !ok {
		return nil, ErrWrongType
	}

	set2, ok := second.asSet()
	if !ok {
		return nil, ErrWrongType
	}

	firstKeys := set1.Items()
	inter := []string{}
	for _, k := range firstKeys {
		if set2.Has(k) {
			inter = append(inter, k)
		}
	}

	return inter, nil
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
