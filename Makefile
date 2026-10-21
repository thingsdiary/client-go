.PHONY: install-tools fmt

install-tools:
	go install tool

fmt:
	go tool gimps .
