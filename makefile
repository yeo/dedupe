CWD := $(shell pwd)

e2e:
	go run main.go $(CWD)/sample -extension="jpg"
