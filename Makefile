export REDIS_ADDRESS="localhost:6379"
export MONGO_ADDRESS="mongodb://localhost:27017/?replicaSet=rs0"

build-runner:
	go build -o ./bin/runner cmd/runner/main.go

run-runner: build-runner
	go run cmd/runner/main.go

build-controlplane:
	go build -o ./bin/control-plane cmd/control_plane/main.go

run-controlplane: build-controlplane
	go run cmd/control_plane/main.go


