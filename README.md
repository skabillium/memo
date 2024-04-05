# Memo

Memo is a Redis-like in-memory database implemented in go. Other than basic key-value functionalities
it supports additional data structures like priority queues, FIFO queues and stacks.

Run the server with `make build && ./bin/memo`, this should start a memo server on `localhost:5678`

## List of supported commands
- PING
- HELLO
- INFO
- AUTH
- FLUSHALL
- KEYS
- EXPIRE
- GET
- SET (only EX option supported)
- DEL
- LPUSH
- LPOP
- RPUSH
- RPOP
- LLEN

### Additional commands not supported by Redis
- VERSION: Prints the version of the server
- CLEANUP: Clean up expired keys

Memo also has support for the priority queue data type for
with the following commands:
- QADD: Add elements to a queue
- QPOP: Remove elements from a queue
- QLEN: Get number of queued elements
