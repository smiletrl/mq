db-start:
	- docker compose up -d

db-restart:
	- docker compose down && docker compose up -d
