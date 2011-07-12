include $(GOROOT)/src/Make.inc

TARG=upstream
GOFILES=upstream.go crawler.go
GOFMT=gofmt -l -w

include $(GOROOT)/src/Make.cmd

test: clean all
	#./${TARG} bash
	./${TARG} gnome-shell
	#./${TARG} crawl

format:
	    ${GOFMT} .

