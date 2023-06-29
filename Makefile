NAME?=heplify-server

PKGLIST=$(shell go list ./... | grep -Ev '/vendor|/metric|/config|/sipparser/internal')


all:
	go build -ldflags "-s -w" -o $(NAME) cmd/heplify-server/*.go

static:
	CGO_ENABLED=1 GOOS=linux CGO_LDFLAGS="-lm -ldl" go build -a -ldflags '-extldflags "-static"' -tags netgo -installsuffix netgo -o $(NAME) cmd/heplify-server/*.go

debug:
	go build -o $(NAME) cmd/heplify-server/*.go

test:
	go vet $(PKGLIST)
	go test $(PKGLIST) -race

.PHONY: clean
clean:
	rm -fr $(NAME)
