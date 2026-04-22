BIN     := define
INSTALL := $(HOME)/.local/bin/$(BIN)

.PHONY: build install uninstall test clean

build:
	go build -o $(BIN) .

install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BIN) $(INSTALL)
	@echo "installed: $(INSTALL)"

uninstall:
	rm -f $(INSTALL)
	@echo "removed: $(INSTALL)"

test:
	go test ./...

clean:
	rm -f $(BIN)
