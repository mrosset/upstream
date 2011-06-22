include $(GOROOT)/src/Make.inc

TARG=watch
GOFILES=watch.go
GOFMT=gofmt -l -w

include $(GOROOT)/src/Make.cmd

test: format all
	./${TARG} -t
	./${TARG} bash
	./${TARG} git

format:
	${GOFMT} .
