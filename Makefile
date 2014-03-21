.PHONY: all clean nuke

all: editor
	make todo

clean:
	go clean

editor:
	go fmt
	go install

nuke:
	go clean -i
