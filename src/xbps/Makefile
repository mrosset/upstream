include $(GOROOT)/src/Make.inc

TARG=xbps
GOFILES=xbps.go masterdir.go common.go template.go
GOFMT=gofmt -l -w

include $(GOROOT)/src/Make.pkg

format:
	${GOFMT} .
