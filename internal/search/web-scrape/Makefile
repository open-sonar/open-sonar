GO := go
BIN := scraper
SRC := scraper.go
PKG := scraper

# dependencies
DEPS := github.com/PuerkitoBio/goquery github.com/go-shiori/go-readability

all: build

init:
	$(GO) mod init $(PKG)
	$(GO) get $(DEPS)

build: $(SRC)
	$(GO) build -o $(BIN) $(SRC)

run:
	$(GO) run $(SRC) -query "$(QUERY)" -pages "$(PAGES)"

clean:
	rm -f $(BIN)

fmt:
	$(GO) fmt $(PKG)

deps:
	$(GO) get -u $(DEPS)

.PHONY: all init build run clean fmt deps
