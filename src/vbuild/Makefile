include $(GOROOT)/src/Make.inc

TARG=vbuild
GOFILES=main.go
GOFMT=gofmt -l -w

include $(GOROOT)/src/Make.cmd

test: clean all

format:
	    ${GOFMT} .

