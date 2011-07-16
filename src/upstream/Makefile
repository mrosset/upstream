include $(GOROOT)/src/Make.inc

export XBPS_SRCPKGDIR:=/home/strings/github/vanilla/srcpkgs

TARG=upstream
GOFILES=upstream.go crawler.go
GOFMT=gofmt -l -w

include $(GOROOT)/src/Make.cmd

test: clean all
	./${TARG} check gnome-shell

format:
	    ${GOFMT} .

