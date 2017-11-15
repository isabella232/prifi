.PHONY: all
all: test coveralls it it2

.PHONY: test_fmt
test_fmt:
	@echo Checking correct formatting of files...
	@{ \
		files=$$( go fmt ./... ); \
		if [ -n "$$files" ]; then \
		echo "Files not properly formatted: $$files"; \
		exit 1; \
		fi; \
	}

.PHONY: test_fmt
build:
	@echo Testing build...
	@{ \
		go build sda/app/prifi.go && rm -f prifi; \
	}

.PHONY: test_govet
test_govet:
	@echo Running go vet...
	@{ \
		if ! go vet ./...; then \
		exit 1; \
		fi \
	}

.PHONY: coveralls
coveralls:
	./coveralls.sh

.PHONY: test_verbose
test_verbose:
	go test -v -race -short ./...

.PHONY: it
it:
	./test.sh integration || cat relay.log

.PHONY: it2
it2:
	./test.sh integration2 || cat relay.log

.PHONY: clean
clean:
	rm -f profile.cov *.log timing.txt prifi-lib/relay/timing.txt

.PHONY: test
test: build test_fmt test_govet

.PHONY: test_lint
test_lint:
	@echo Checking linting of files ...
	@{ \
		go get -u github.com/golang/lint/golint; \
		exclude="_test.go|ALL_CAPS|underscore|should be of the form|.deprecated|and that stutters|error strings should not be capitalized"; \
		lintfiles=$$( golint ./... | egrep -v '($$exclude)' ); \
		if [ -n "$$lintfiles" ]; then \
		echo "Lint errors:"; \
		echo "$$lintfiles"; \
		exit 1; \
		fi \
	}