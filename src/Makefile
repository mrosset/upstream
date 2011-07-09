GOFMT=gofmt -l -w
DIRS=xbps upstream vbuild shhelper

clean.dirs: $(addsuffix .clean, $(DIRS))
install.dirs: $(addsuffix .install, $(DIRS))
test.dirs: $(addsuffix .test, $(DIRS))

install: $(patsubst %,%.install,$(DIRS))
clean: $(patsubst %,%.clean,$(CLEANDIRS))

%.test:
	+@echo test $*
	+@$(MAKE) -C $* install
	+@$(MAKE) -C $* test


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

clean: 		clean.dirs

install: 	install.dirs

test: 		test.dirs

format:
	${GOFMT} .
