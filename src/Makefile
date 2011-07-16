DIRS=xbps upstream gosrcpkg vbuild

export XBPS_SRCPKGDIR:=/home/strings/github/vanilla/srcpkgs

all: $(patsubst %,%.all,$(DIRS))
test: $(patsubst %,%.test,$(DIRS))
clean: $(patsubst %,%.clean,$(DIRS))
format: $(patsubst %,%.format,$(DIRS))

%.all:
	 $(MAKE) -C $* all

%.test:
	+@echo test $*
	+@$(MAKE) -C $* test 2>&1 || echo TEST FAIL $*
%.clean:
	$(MAKE) -C $* clean

%.format:
	 $(MAKE) -C $* format
