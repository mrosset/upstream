include $(GOROOT)/src/Make.inc

TARG=upstream
GOFILES=upstream.go template.go
GOFMT=gofmt -l -w

include $(GOROOT)/src/Make.cmd

test: format all
	./${TARG} -test

format:
	${GOFMT} .
