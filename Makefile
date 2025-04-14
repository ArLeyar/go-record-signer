.PHONY: init sign test

init:
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example"; \
		cp .env.example .env; \
	fi
	docker-compose down -v
	docker-compose build
	docker-compose up initdb

sign:
	docker-compose up dispatcher worker

test:
	go test ./... 