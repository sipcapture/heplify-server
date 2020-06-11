NAME?=heplify-server

PKGLIST=$(shell go list ./... | grep -Ev '/vendor|/metric|/config|/sipparser/internal')


all:
	@if [ "` ldconfig -p | grep libluajit-5.1 $$f`" = "" ]; then\
		sudo apt-get install -y luajit-5.1;\
	fi
	go build -ldflags "-s -w" -o $(NAME) cmd/heplify-server/*.go

debug:
	go build -o $(NAME) cmd/heplify-server/*.go

test:
	go vet $(PKGLIST)
	go test $(PKGLIST) -race

.PHONY: clean
clean:
	rm -fr $(NAME)
