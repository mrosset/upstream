GOFMT=gofmt -l -w

all:
	make -C ./cmd
	make -C ./pkg

clean:
	make -C ./cmd clean
	make -C ./pkg clean

install:
	make -C ./pkg install
	make -C ./cmd install

format:
	${GOFMT} .
