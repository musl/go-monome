.PHONY: all clean

all: clean test cmds

clean:
	go clean ./...

test: 
	go test -v .

cmds:
	for d in cmd/* ;\
		do \
		make -C $$d ;\
		done

