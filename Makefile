NAME?=heplify-server

PKGLIST=$(shell go list ./... | grep -Ev '/vendor|/metric|/config|/sipparser')

all:
	go build -ldflags "-s -w" -o $(NAME) cmd/heplify-server/*.go

debug:
	go build -o $(NAME) cmd/heplify-server/*.go

test:
	go vet $(PKGLIST)
	go test $(PKGLIST)

.PHONY: clean
clean:
	rm -fr $(NAME)
