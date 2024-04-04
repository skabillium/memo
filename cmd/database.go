package main

import "errors"

type DataStoreType = byte

const (
	DsValue DataStoreType = iota
	DsQueue
	DsPQueue
	DsList
)

type DataStore struct {
	Kind   DataStoreType
	Value  string
	Queue  *Queue
	PQueue *PriorityQueue
	List   *List
}

type Database struct {
	stores map[string]*DataStore
}

func NewDatabase() *Database {
	return &Database{stores: make(map[string]*DataStore)}
}

func (d *Database) FlushAll() {
	d.stores = make(map[string]*DataStore)
}

func (d *Database) Keys() []string {
	// TODO: Pre-allocate instead of appending
	keys := []string{}
	for k := range d.stores {
		keys = append(keys, k)
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

func (d *Database) QPop(qname string) (string, bool, error) {
	store, found := d.stores[qname]
	if !found {
		return "", false, nil
	}

	if store.Kind != DsQueue {
		return "", found, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
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

func (d *Database) PQPop(qname string) (string, bool, error) {
	store, found := d.stores[qname]
	if !found {
		return "", found, nil
	}

	if store.Kind != DsPQueue {
		return "", found, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
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

func (d *Database) LPush(lname string, value string) {
	store, found := d.stores[lname]
	if !found {
		store = d.createList()
	}

	store.List.Prepend(value)
	if !found {
		d.stores[lname] = store
	}
}

func (d *Database) RPush(lname string, value string) {
	store, found := d.stores[lname]
	if !found {
		store = d.createList()
	}

	store.List.Append(value)
	if !found {
		d.stores[lname] = store
	}
}

func (d *Database) LPop(lname string) (string, bool, error) {
	store, found := d.stores[lname]
	if !found {
		return "", found, nil
	}

	if store.Kind != DsList {
		return "", found, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	value := store.List.PopHead()
	if store.List.Length == 0 {
		d.Del(lname)
	}

	return value, found, nil
}

func (d *Database) RPop(lname string) (string, bool, error) {
	store, found := d.stores[lname]
	if !found {
		return "", found, nil
	}

	if store.Kind != DsList {
		return "", found, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	value := store.List.PopTail()
	if store.List.Length == 0 {
		d.Del(lname)
	}

	return value, found, nil
}

func (d *Database) LLen(lname string) (int, error) {
	store, found := d.stores[lname]
	if !found {
		return -1, nil
	}

	if store.Kind != DsList {
		return -1, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return store.List.Length, nil
}

func (d *Database) createValue(value string) *DataStore {
	return &DataStore{Kind: DsValue, Value: value}
}

func (d *Database) createQueue() *DataStore {
	return &DataStore{Kind: DsQueue, Queue: NewQueue()}
}

func (d *Database) createPQueue() *DataStore {
	return &DataStore{Kind: DsPQueue, PQueue: NewPriorityQueue()}
}

func (d *Database) createList() *DataStore {
	return &DataStore{Kind: DsList, List: NewList()}
}
