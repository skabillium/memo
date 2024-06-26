# Memo

Memo is a Redis-compliant in-memory database implemented in go. Other than basic key-value 
functionalities it also supports lists, priority queues and sets. Like Redis, it also
supports the RESP protocol so it can be used with any Redis client library in your programming
language of choice.

Run the server with `make build && ./bin/memo`, this should start a memo server on `localhost:5678`

## Running the project
Run `make start` to start the server at `localhost:5678` with default authentication:
- User: memo
- Password: password

If you want to enable the Write Ahead log for persisting data between restarts, run the server
with the `--wal` option. This will persist all queries to the `wal.log` file which will be read
and executed in the next restart.

For a complete list of supported CLI options run `make help`.

## List of supported commands
- `QUIT`
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
