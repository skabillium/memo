# Memo

Memo is a Redis-compliant in-memory database implemented in go. Other than basic key-value 
functionalities it also supports lists, priority queues and sets.

Run the server with `make build && ./bin/memo`, this should start a memo server on `localhost:5678`

## List of supported commands
- `PING`
- `HELLO`
- `INFO`
- `DBSIZE`
- `AUTH`
- `FLUSHALL`
- `KEYS`
- `EXPIRE`
- `GET`
- `SET` (only EX option supported)
- `DEL`
- `LPUSH`
- `LPOP`
- `RPUSH`
- `RPOP`
- `LLEN`
- `SADD`
- `SISMEMBER`
- `SREM`
- `SCARD`
- `SINTER`

### Additional commands not supported by Redis
- `VERSION`: Prints the version of the server
- `CLEANUP`: Clean up expired keys

Memo also has support for the priority queue data type for
with the following commands:
- `QADD key element [element...]`: Add elements to a queue
- `QPOP key`: Remove element from a queue
- `QLEN key`: Get number of queued elements

## Running the test suite
To run the unit test suite for the database internals run `make tests`. If you instead want to run
the integration test suit run the following commands:
```sh
make install # Download dependencies for integration tests

make build
./bin/memo -noauth # Start server without authentication

make integ # Run integration test suite
```
