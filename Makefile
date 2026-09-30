.PHONY: up down logs ps psql clean migrate-up migrate-down

up:
	docker compose up -d --wait

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

psql:
	docker compose exec postgres psql -U loanflow -d loanflow

clean:
	docker compose down -v

migrate-up:
	docker compose run --rm migrate up

migrate-down:
	docker compose run --rm migrate down 1
