# Cómo trabajamos en este repositorio

La guía completa, con ejemplos, está en [`docs/plan/`](docs/plan/). Acá va el resumen para tener a mano.

## 1. Ramas

`main` está protegida: sólo se modifica mediante Pull Request aprobado por otro integrante.

| Tipo | Nombre de la rama |
|---|---|
| Historia de usuario | `feature/HU-01-crear-proyecto` |
| Ítem técnico | `chore/TEC-02-ci` |
| Defecto | `fix/DEF-4-descripcion-corta` |
| Documentación | `docs/sprint-1-retro` |

```bash
git switch main
git pull
git switch -c feature/HU-01-crear-proyecto
```

## 2. Commits (evidencia de TDD)

Usamos [Conventional Commits](https://www.conventionalcommits.org/es/) con el ID de la historia y la fase de TDD:

```text
test(HU-01): RED - nombre con menos de 3 caracteres es inválido
feat(HU-01): GREEN - valida largo del nombre
refactor(HU-01): extrae normalización del nombre
docs(HU-01): spec SDD y escenarios BDD
```

Un commit por fase: primero el test que falla (**RED**), después el código mínimo que lo hace pasar (**GREEN**)
y por último la mejora sin romper los tests (**REFACTOR**). Los profesores evalúan TDD mirando este historial.

## 3. Pull Requests

1. `git push -u origin <rama>` y `gh pr create` (o desde la web).
2. Completar la plantilla del PR, incluida la trazabilidad.
3. Revisión cruzada obligatoria: **Juan Pablo revisa a Mariano, Mariano revisa a Carolina, Carolina revisa a Juan Pablo**.
4. Mergear con **"Create a merge commit"** o **"Rebase and merge"**. *Squash* está deshabilitado porque borraría los commits RED/GREEN/REFACTOR.
5. Mover la tarjeta a **Hecho** en el tablero.

## 4. Definition of Ready y Definition of Done

Ver [`docs/scrum/definition-of-ready-done.md`](docs/scrum/definition-of-ready-done.md).

## 5. Uso de IA

Está permitido, pero cada integrante debe entender y poder explicar todo lo que sube.
Registrar qué se pidió, qué se usó y cómo se validó en [`docs/ia/registro.md`](docs/ia/registro.md).

## 6. Secretos

Las claves de Supabase van en `.env` (copiar desde `.env.example`). **Nunca** subir `.env` al repositorio.

## 7. Base de datos y migraciones

El esquema de la base (Supabase) se arma con migraciones SQL versionadas en [`migrations/`](migrations/), que se aplican con goose
desde un comando propio. Una migración que ya se aplicó **nunca se edita**: cualquier cambio es una migración nueva
(`0002_...sql`) y se actualiza [`docs/modelo-datos.md`](docs/modelo-datos.md).

| Quiero... | Comando |
|---|---|
| Ver qué migraciones están aplicadas | `go run ./cmd/migrar status` |
| Aplicar las que faltan | `go run ./cmd/migrar up` |
| Deshacer todo (borra los datos; solo en bases de desarrollo) | `go run ./cmd/migrar down --borrar-todo` |

El comando usa `MIGRATIONS_DATABASE_URL` o, si no está, `DATABASE_URL` (ver `.env.example`). **Nunca** apuntarlo a producción
sin que el equipo lo haya decidido.

**Tests con base de datos.** Los tests de `internal/store/postgres`, `app` y `cmd/migrar` necesitan un PostgreSQL descartable
y se omiten si no existe `TEST_DATABASE_URL`. Con Docker Desktop abierto:

```bash
docker run -d --name seguimiento-pg-test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=seguimiento -p 54329:5432 postgres:16-alpine
export TEST_DATABASE_URL='postgres://postgres:test@localhost:54329/seguimiento?sslmode=disable'
go test ./...
```

En PowerShell, la segunda línea es `$env:TEST_DATABASE_URL = 'postgres://postgres:test@localhost:54329/seguimiento?sslmode=disable'`.
Cada test crea su propio esquema y lo borra al terminar, así que no ensucian la base. El CI levanta su propio Postgres.
