include $(GOROOT)/src/Make.inc

export XBPS_SRCPKGDIR:=/home/strings/vanilla/srcpkgs

TARG=xbps
GOFILES=xbps.go masterdir.go template.go common.go
GOFMT=gofmt -l -w
include $(GOROOT)/src/Make.pkg

format:
	    ${GOFMT} .

