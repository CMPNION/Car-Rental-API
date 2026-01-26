SHELL := /bin/sh

.PHONY: dev backend-dev frontend-dev frontend-install

backend-dev:
	go run ./cmd/main/main.go

frontend-install:
	cd frontend/ && npm install

frontend-dev: frontend-install
	cd frontend/ && npm run dev

dev:
	$(MAKE) -j2 backend-dev frontend-dev
