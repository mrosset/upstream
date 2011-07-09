include $(GOROOT)/src/Make.inc

TARG=shhelper
GOFILES=main.go
GOFMT=gofmt -l -w

include $(GOROOT)/src/Make.cmd

test: install
	xbps-src remove wget
	xbps-src -C -K install wget

format:
	${GOFMT} .

install: clean format all
	cp -f ${TARG} ${HOME}/masterdir/usr/local/bin
