build:
	go build -o ./tmp/bin/app ./cmd/main.go

run: build
	./tmp/bin/app

clean:
	rm -f -r ./tmp
	rm -f -r ./VOLUMES

up-env:
	docker compose up -d

down-env:
	docker compose down