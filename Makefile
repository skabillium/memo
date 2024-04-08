CMD = ./cmd
TEST = ./test
BIN = ./bin/memo

clean:
	rm -rf ./bin

test:
	go test ${CMD}/db
	go test ${CMD}/resp
	go test ${CMD}

testsuite:
	go test ${TEST}

dev:
	go run ${CMD} --noauth

build:
	go build -o ${BIN} ${CMD}
