BIN     := define
INSTALL := $(HOME)/.local/bin/$(BIN)
LABS    := labs/repos
REPORTS := labs/reports

.PHONY: build install uninstall test clean labs-fetch labs-analyze labs-phpunit-demo labs-clean

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

# Clone / update all repos listed in labs/repos.txt
labs-fetch:
	@mkdir -p $(LABS)
	@grep -v '^\s*#' labs/repos.txt | grep -v '^\s*$$' | while read name url branch profile; do \
		dest=$(LABS)/$$name; \
		if [ -d "$$dest/.git" ]; then \
			echo "→ update  $$name"; \
			git -C "$$dest" pull --quiet; \
		else \
			echo "→ clone   $$name ($$branch)"; \
			git clone --quiet --depth=1 --branch "$$branch" "$$url" "$$dest" 2>&1 | grep -v '^warning'; \
		fi; \
	done

# Run define extract + check on every repo, using the profile from labs/profiles/
# Profile lives outside the clone so experiments don't touch the upstream code.
labs-analyze: build
	@mkdir -p $(REPORTS)
	@grep -v '^\s*#' labs/repos.txt | grep -v '^\s*$$' | while read name url branch profile; do \
		dir=$(LABS)/$$name; \
		cfg=labs/profiles/$$profile.yml; \
		out=$(REPORTS)/$$name.def; \
		report=$(REPORTS)/$$name.txt; \
		[ -d "$$dir" ] || { echo "skip $$name (not cloned)"; continue; }; \
		echo ""; \
		echo "=== $$name  [profile: $$profile] ==="; \
		./$(BIN) extract "$$dir" --config "$$cfg" -o "$$out" 2>/dev/null && \
		./$(BIN) check "$$out" 2>&1 | tee "$$report" || true; \
	done

# PHPUnit demo: compare extraction WITHOUT vs WITH tests.
# Shows that a library's public API looks "dead" until you include its consumers (tests).
labs-phpunit-demo: build
	@mkdir -p $(REPORTS)
	@dir=$(LABS)/phpunit; \
	[ -d "$$dir" ] || { echo "run 'make labs-fetch' first"; exit 1; }; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo " PHPUnit — src/ only  (library without its consumers)"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	./$(BIN) extract "$$dir" --config labs/profiles/php-phpunit.yml \
		-o $(REPORTS)/phpunit-notests.def 2>/dev/null; \
	./$(BIN) check $(REPORTS)/phpunit-notests.def "TextUI/Application" 2>&1 | tee $(REPORTS)/phpunit-notests.txt; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo " PHPUnit — src/ + tests/  (library with its consumers)"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	./$(BIN) extract "$$dir" --config labs/profiles/php-phpunit-with-tests.yml \
		-o $(REPORTS)/phpunit-tests.def 2>/dev/null; \
	./$(BIN) check $(REPORTS)/phpunit-tests.def "TextUI/Application" 2>&1 | tee $(REPORTS)/phpunit-tests.txt

labs-clean:
	rm -rf $(LABS) $(REPORTS)/*.def $(REPORTS)/*.txt
