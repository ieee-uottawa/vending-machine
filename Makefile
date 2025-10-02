.PHONY: server tui

server:
	go build -o vm-server ./cmd/server

tui:
	go build -o vm-tui ./cmd/tui