.PHONY: all install build test test_fmt test_govet test_verbose test_lint coveralls it it2 it-verbose it2-verbose clean
all: test coveralls it-verbose it2-verbose

install:
	./prifi.sh install

test: install build test_fmt test_govet test_lint
	DEBUG_COLOR="True" DEBUG_LVL=1 go test -short ./...

test_fmt:
	@echo Checking correct formatting of files...
	@{ \
		files=$$( go fmt ./... ); \
		if [ -n "$$files" ]; then \
		echo "Files not properly formatted: $$files"; \
		exit 1; \
		fi; \
	}

build:
	@echo Testing build...
	@{ \
		go build sda/app/prifi.go && rm -f prifi; \
	}

test_govet:
	@echo Running go vet...
	@{ \
		if ! go vet ./...; then \
		exit 1; \
		fi \
	}

coveralls:
	./coveralls.sh

test_verbose:
	DEBUG_COLOR="True" DEBUG_LVL=3 go test -v -race ./...

it:
	./test.sh integration

it2:
	./test.sh integration2

it-verbose:
	./test.sh integration || (cat relay.log; exit 1)

it2-verbose:
	./test.sh integration2 || (cat relay.log; exit 1)

clean:
	rm -f profile.cov *.log timing.txt prifi-lib/relay/timing.txt

test_lint:
	@echo Checking linting of files ...
	@{ \
		go get -u golang.org/x/lint/golint; \
		exclude="_test.go|ALL_CAPS|underscore|should be of the form|.deprecated|and that stutters|error strings should not be capitalized"; \
		lintfiles=$$( golint ./... | egrep -v "($$exclude)" ); \
		if [ -n "$$lintfiles" ]; then \
		echo "Lint errors:"; \
		echo "$$lintfiles"; \
		exit 1; \
		fi \
	}
