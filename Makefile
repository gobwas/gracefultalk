
all: A B C

A:
	go build -o bin/client ./src/example/A/cmd/client
	go build -o bin/server ./src/example/A/cmd/server

B:
	go build -o bin/twin ./src/example/B/cmd/twin

C:
	go build -o bin/master ./src/example/C/cmd/master
	go build -o bin/worker ./src/example/C/cmd/worker
