CMD = ./cmd
BIN = ./bin/memo

clean:
	rm -rf ./bin

test:
	go test ${CMD}/db
	go test ${CMD}/resp
	go test ${CMD}

dev:
	go run ${CMD} --noauth

build:
	go build -o ${BIN} ${CMD}
