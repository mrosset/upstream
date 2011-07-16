include $(GOROOT)/src/Make.inc

export XBPS_SRCPKGDIR:=/home/strings/github/vanilla/srcpkgs

TARG=gosrcpkg
GOFILES=main.go
GOFMT=gofmt -l -w

include $(GOROOT)/src/Make.cmd

format:
	    ${GOFMT} .

test: clean format all
	./${TARG} json bash
	./${TARG} shell bash
	./${TARG} chkdepends fontconfig
