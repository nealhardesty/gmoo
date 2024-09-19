default: build

NAME=gmoo

build:
	mkdir -p bin
	go build -o bin/$(NAME)

clean:
	rm -f bin/$(NAME)

run:
	go run . 

