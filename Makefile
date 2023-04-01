lint:
	golangci-lint run -c ./golangci.yml ./...

test:
	go test `go list ./... | grep -v ./examples` -v --cover

test-report:
	go test `go list ./... | grep -v ./examples` -v --cover -covermode=atomic -coverprofile=coverage.out
	go tool cover -html=coverage.out

run_base_example:
	cd examples/base && go run main.go serve
