app.docker.test.start:
	docker compose --file ./tests/config/docker-compose.yml  up -d

test: app.docker.test.start
		go test -count=1 -tags integration  ./tests/integration/

test.verbose: app.docker.test.start
		go test -count=1 -tags integration -v  ./tests/integration/

app.docker.start:
	docker compose --file ./docker-compose.yml  up -d

run: app.docker.start 
		go run cmd/main.go 
