.PHONY: all clean

all: clean test 

clean:
	go clean ./...

test:
	go test

