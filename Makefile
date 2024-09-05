build:
	GOOS=linux CGO_ENABLED=0 go build -o ./build/go-pgctl ./cmd/

run:
	go run ./cmd/

rs_up: build
	cd ./deployment && docker compose -f docker-compose-rs.yaml up -d

rs_build_up: build
	cd ./deployment && docker compose -f docker-compose-rs.yaml --build up -d

rs_down:
	cd ./deployment && docker compose -f docker-compose-rs.yaml down

zk_up:
	cd ./deployment && docker compose -f docker-compose-zk.yaml up -d

zk_down:
	cd ./deployment && docker compose -f docker-compose-zk.yaml down

.PHONY: build