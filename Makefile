NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

TEST?=$(shell go list ./... | grep -v vendor)

# Get the current full sha from git
GITSHA:=$(shell git rev-parse HEAD)

# Get the current local branch name from git (if we can, this may be blank)
GITBRANCH:=$(shell git symbolic-ref --short HEAD 2>/dev/null)

all:
	@mkdir -p bin/
	@printf "$(OK_COLOR)==> Building$(NO_COLOR)\n"
	@go build -ldflags "-X github.com/masterzen/dashlane-cli/version.GitSHA=${GITSHA}" -o bin/dashlane-cli .

clean:
	@rm -rf bin/ pkg/ src/

format:
	go fmt `go list ./... | grep -v vendor`

ci:
	@printf "$(OK_COLOR)==> Testing with Coveralls...$(NO_COLOR)\n"
	"$(CURDIR)/scripts/test.sh"

test:
	@printf "$(OK_COLOR)==> Testing...$(NO_COLOR)\n"
	@go test $(TEST) $(TESTARGS) -timeout=2m

.PHONY: all clean
