include $(GOROOT)/src/Make.inc

TARG=watch
GOFILES=watch.go
GOFMT=gofmt -l -w

include $(GOROOT)/src/Make.cmd

test: format all
	./${TARG} -t

format:
	${GOFMT} .
