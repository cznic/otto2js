# Copyright 2014 The otto2js Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

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
