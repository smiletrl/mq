db-start:
	- docker compose up -d
	- cd cmd/init && go run main.go

db-restart:
	- docker compose down && docker compose up -d

app-start:
	- cd cmd/api && go run main.go
