include $(GOROOT)/src/Make.inc

TARG=shhelper
GOFILES=main.go
GOFMT=gofmt -l -w

include $(GOROOT)/src/Make.cmd

format:
	    ${GOFMT} .

test:
	./${TARG} json bash
	./${TARG} shell bash
