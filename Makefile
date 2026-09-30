.PHONY: up down logs ps psql clean

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
