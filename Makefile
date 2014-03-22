.PHONY: all clean nuke todo

all: editor
	make todo

clean:
	go clean

editor:
	go fmt
	go install

nuke:
	go clean -i

todo:
	grep -n TODO *
