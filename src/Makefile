GOFMT=gofmt -l -w
DIRS=xbps vbuild

clean.dirs: $(addsuffix .clean, $(DIRS))
install.dirs: $(addsuffix .install, $(DIRS))
test.dirs: $(addsuffix .test, $(DIRS))

install: $(patsubst %,%.install,$(DIRS))
clean: $(patsubst %,%.clean,$(CLEANDIRS))

%.test:
	+@echo test $*
	+@$(MAKE) -C $* test.clean >$*/test.out 2>&1 || (echo TEST FAIL $*; cat $*/test.out; exit 1)


echo-dirs:
	@echo $(DIRS)

%.install: %.clean
	@echo
	@echo %%%% making $* %%%%
	@echo
	$(MAKE) -C $* install

%.clean:
	@echo
	@echo %%%% cleaning $* %%%%
	@echo
	$(MAKE) -C $* clean

clean: clean.dirs

install: install.dirs

test:	test.dirs

format:
	${GOFMT} .
