CMD = ./cmd
TEST = ./test
BIN = ./bin/memo

install:
	go mod download

clean:
	rm -rf ./bin

tests:
	go test ${CMD}/db
	go test ${CMD}/resp
	go test ${CMD}

integ:
	go test ${TEST}

dev:
	go run ${CMD} --noauth

build:
	go build -o ${BIN} ${CMD}

prod: clean build
	${BIN}
