.PHONY: init dispatch sign check test  

init:
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example"; \
		cp .env.example .env; \
	fi
	docker-compose down -v --remove-orphans
	docker-compose build --no-cache
	docker-compose up initdb

dispatch:
	docker-compose up dispatcher

sign:
	docker-compose up dispatcher worker1 worker2

check:
	@echo "Checking database status..."
	@chmod +x ./check_db.sh
	@./check_db.sh

test:
	go test ./... 