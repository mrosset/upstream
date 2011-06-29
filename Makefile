include $(GOROOT)/src/Make.inc

TARG=upstream
GOFILES=upstream.go
GOFMT=gofmt -l -w

include $(GOROOT)/src/Make.cmd

test: format all
	./${TARG} -c
	./${TARG} bash
	./${TARG} git
	./${TARG} -b
	./${TARG} -home ncdu
	./${TARG} -ct

format:
	${GOFMT} .
