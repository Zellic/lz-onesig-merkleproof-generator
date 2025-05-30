.PHONY: build clean

BINARY_NAME=merkle-cli

build:
	go build -o $(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
