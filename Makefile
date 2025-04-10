.PHONY: init start

init:
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example"; \
		cp .env.example .env; \
	fi
	docker-compose down -v
	docker-compose build
	docker-compose up initdb

start:
	docker-compose up 