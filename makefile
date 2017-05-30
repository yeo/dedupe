CWD := $(shell pwd)

build:
	go build -o dedupe .
e2e:
	go run main.go -extension jpg $(CWD)/sample
run:
	git co sample
	rm -rf output/*.jpg
	go run main.go -extension jpg -move-to="./output" $(CWD)/sample 
