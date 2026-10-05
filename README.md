# Software Metrics & Estimation

Aplicación web para **estimar, planificar, seguir y medir proyectos de software**: Product Backlog, Sprints,
Planning Poker, registro de esfuerzo, defectos, métricas, dashboard y reportes.

Trabajo Práctico Integrador de **Ingeniería y Calidad de Software 2026** — UTN Facultad Regional San Rafael.

**App publicada:** https://seguimiento-medicion.vercel.app — cada merge a `main` se publica solo; cada Pull Request
tiene su propio preview (link en el comentario de Vercel del PR).

## Equipo

| Integrante | GitHub | Rol Scrum | Responsable de |
|---|---|---|---|
| López, Juan Pablo | @Juan-Pablo-Lopez-Laszuk | Agile Enabler · Product Builder | E1 Proyectos · E3 Sprints · E9 Reportes · Arquitectura y CI |
| Piastrellini, Mariano | @MarianoPiastre | Product Builder | E2 Product Backlog · E4 Estimación y Planning Poker · Base de datos · Login |
| Bravo, Carolina | @CaritoNBravo | Product Builder | E5 Esfuerzo · E6 Defectos · E7 Métricas · E8 Dashboard · Deploy · Diseño |

Product Architect (Cliente): profesores de la cátedra.

## Stack

- **Go** — dominio, reglas de negocio, validaciones y servidor web (`net/http` + `chi`).
- **HTMX + Bootstrap 5 + Chart.js** — interfaz servida desde Go con `html/template`.
- **Supabase** (PostgreSQL + Auth) — datos y login.
- **Vercel** — deploy automático desde GitHub (producción = `main`, preview por cada PR).
- **godog** — escenarios BDD en español; `go test` para TDD.

## Metodología

Cada funcionalidad sigue la cadena de trazabilidad:

**Historia de Usuario → Especificación SDD → Criterios de Aceptación → Escenarios BDD → Tests → Código Go**

| Artefacto | Dónde |
|---|---|
| Historias de usuario | Issues + [tablero del proyecto](https://github.com/users/Juan-Pablo-Lopez-Laszuk/projects) |
| Especificaciones SDD | [`specs/`](specs/) |
| Escenarios BDD | [`features/`](features/) |
| Ceremonias Scrum y actas | [`docs/scrum/`](docs/scrum/) |
| Registro de uso de IA | [`docs/ia/registro.md`](docs/ia/registro.md) |
| Plan de trabajo y guía | [`docs/plan/`](docs/plan/) |

Cómo trabajar en el repositorio (ramas, commits, Pull Requests): ver [CONTRIBUTING.md](CONTRIBUTING.md).

## Sprints

| Sprint | Fechas | Objetivo |
|---|---|---|
| 0 · Preparación | 28/09 – 04/10 | Entorno, tablero, Product Backlog inicial |
| 1 · MVP | 05/10 – 11/10 | Versión funcional básica |
| 2 · Interfaz | 12/10 – 18/10 | Interfaz usable, login y Planning Poker |
| 3 · Funcionalidad y Calidad | 19/10 – 25/10 | Dashboard, métricas completas, reportes, robustez |
| 4 · Cierre | 26/10 – 01/11 | Exportación a PDF, documentación y entrega final |

## Cómo correrlo

Requisitos: [Go](https://go.dev/dl/) 1.24 o superior y Git.

```bash
go run ./cmd/server          # levanta la app en http://localhost:8080 (salud: /health)
go test ./...                # tests unitarios + escenarios BDD
go test ./features/... -v    # solo los escenarios BDD (godog)
go test -cover ./...         # tests con porcentaje de cobertura
```

Con `make` instalado también están `make run`, `make test`, `make cover`, `make bdd` y `make lint`.

## Estructura

```text
api/index.go          función de Vercel: delega cada pedido al router (a través de app/)
app/                  puerta de entrada pública al router (Vercel no puede importar internal/)
cmd/server/           servidor local
internal/server/      router, handlers y renderizado de páginas
internal/domain/      entidades y reglas de negocio (un subpaquete por épica)
internal/service/     casos de uso
internal/metrics/     cálculos de métricas
internal/store/       repositorios: memory (tests) y postgres (Supabase)
web/templates/        layout base (base.html) y una plantilla por página (paginas/)
features/             escenarios BDD (.feature) y sus pasos en Go
specs/                especificaciones SDD
docs/                 métricas, Scrum, registro de IA, plan y guía
```
