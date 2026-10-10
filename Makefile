# Atajos para las tareas habituales. En Windows sin `make`, usar los comandos de cada regla directamente.

.PHONY: run test cover bdd fmt vet lint db-test migrar migrar-estado

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

db-test: ## Levanta un PostgreSQL descartable en Docker (puerto 54329) para los tests con base de datos
	docker run -d --name seguimiento-pg-test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=seguimiento -p 54329:5432 postgres:16-alpine
	@echo "Ahora definí TEST_DATABASE_URL=postgres://postgres:test@localhost:54329/seguimiento?sslmode=disable"

migrar: ## Aplica las migraciones pendientes a la base de DATABASE_URL (o MIGRATIONS_DATABASE_URL)
	go run ./cmd/migrar up

migrar-estado: ## Muestra qué migraciones están aplicadas
	go run ./cmd/migrar status
