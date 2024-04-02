package main

import "errors"

type DataStoreType = byte

const (
	DsValue DataStoreType = iota
	DsQueue
	DsPQueue
)

type DataStore struct {
	Kind   DataStoreType
	Value  MemoString
	Queue  *Queue
	PQueue *PriorityQueue
}

type Database struct {
	stores map[string]*DataStore
}

func NewDatabase() *Database {
	return &Database{stores: make(map[string]*DataStore)}
}

func (d *Database) Keys() []MemoString {
	keys := []MemoString{}
	for k := range d.stores {
		keys = append(keys, MemoString(k))
	}

	return keys
}

func (d *Database) Get(key string) (*DataStore, bool) {
	store, found := d.stores[key]
	return store, found
}

func (d *Database) Set(key string, value string) {
	d.stores[key] = d.createValue(value)
}

func (d *Database) Del(key string) {
	delete(d.stores, key)
}

func (d *Database) Qadd(qname string, value string) {
	store, found := d.stores[qname]
	if !found {
		store = d.createQueue()
	}

	store.Queue.Enqueue(value)
	if !found {
		d.stores[qname] = store
	}
}

func (d *Database) QPop(qname string) (MemoString, bool, error) {
	store, found := d.stores[qname]
	if !found {
		return MemoString(""), false, nil
	}

	if store.Kind != DsQueue {
		return MemoString(""), found, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return store.Queue.Dequeue(), true, nil
}

func (d *Database) Qlen(qname string) (int, bool, error) {
	store, found := d.stores[qname]
	if !found {
		return -1, found, nil
	}

	if store.Kind != DsQueue {
		return -1, found, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return store.Queue.Length, found, nil
}

func (d *Database) PQAdd(qname string, value string, priority int) error {
	store, found := d.stores[qname]
	if !found {
		store = d.createPQueue()
	}

	if store.Kind != DsPQueue {
		return errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	store.PQueue.Enqueue(value, priority)
	if !found {
		d.stores[qname] = store
	}

	return nil
}

func (d *Database) PQPop(qname string) (MemoString, bool, error) {
	store, found := d.stores[qname]
	if !found {
		return MemoString(""), found, nil
	}

	if store.Kind != DsPQueue {
		return MemoString(""), found, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return store.PQueue.Dequeue(), found, nil
}

func (d *Database) PQLen(qname string) (int, bool, error) {
	store, found := d.stores[qname]
	if !found {
		return -1, found, nil
	}

	if store.Kind != DsPQueue {
		return -1, found, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return store.PQueue.Length, found, nil
}

func (d *Database) createValue(value string) *DataStore {
	return &DataStore{Kind: DsValue, Value: MemoString(value)}
}

func (d *Database) createQueue() *DataStore {
	return &DataStore{Kind: DsQueue, Queue: NewQueue()}
}

func (d *Database) createPQueue() *DataStore {
	return &DataStore{Kind: DsPQueue, PQueue: NewPriorityQueue()}
}
