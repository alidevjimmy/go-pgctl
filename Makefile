build:
	go build -o ./cmd/build ./cmd/

run:
	go run ./cmd/

rs_up:
	cd ./deployment && docker compose -f docker-compose-rs.yaml up -d

rs_down:
	cd ./deployment && docker compose -f docker-compose-rs.yaml down

zk_up:
	cd ./deployment && docker compose -f docker-compose-zk.yaml up -d

zk_down:
	cd ./deployment && docker compose -f docker-compose-zk.yaml down