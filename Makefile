# Atajos para las tareas habituales. En Windows sin `make`, usar los comandos de cada regla directamente.

.PHONY: run test cover bdd fmt vet lint

run: ## Levanta el servidor local en http://localhost:8080
	go run ./cmd/server

test: ## Tests unitarios y escenarios BDD
	go test ./...

cover: ## Tests con reporte de cobertura (abre coverage.html)
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

bdd: ## Solo los escenarios BDD (godog)
	go test ./features/... -v

fmt: ## Formatea el código
	gofmt -w .

vet: ## Análisis estático de Go
	go vet ./...

lint: ## Linter (requiere golangci-lint instalado)
	golangci-lint run ./...
