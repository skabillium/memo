CMD = ./cmd
BIN = ./bin/memo

clean:
	rm -rf ./bin

test:
	go test ${CMD}

dev:
	go run ${CMD}

build:
	go build -o ${BIN} ${CMD}
