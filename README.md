# Software Metrics & Estimation

Aplicación web para **estimar, planificar, seguir y medir proyectos de software**: Product Backlog, Sprints,
Planning Poker, registro de esfuerzo, defectos, métricas, dashboard y reportes.

Trabajo Práctico Integrador de **Ingeniería y Calidad de Software 2026** — UTN Facultad Regional San Rafael.

## Equipo

| Integrante | Rol Scrum | Responsable de |
|---|---|---|
| López, Juan Pablo | Agile Enabler · Product Builder | E1 Proyectos · E3 Sprints · E9 Reportes · Arquitectura y CI |
| Piastrellini, Mariano | Product Builder | E2 Product Backlog · E4 Estimación y Planning Poker · Base de datos · Login |
| Bravo, Carolina | Product Builder | E5 Esfuerzo · E6 Defectos · E7 Métricas · E8 Dashboard · Deploy · Diseño |

Product Architect (Cliente): profesores de la cátedra.

## Stack

- **Go** — dominio, reglas de negocio, validaciones y servidor web (`net/http` + `chi`).
- **HTMX + Bootstrap 5 + Chart.js** — interfaz servida desde Go con `html/template`.
- **Supabase** (PostgreSQL + Auth) — datos y login.
- **Vercel** — deploy automático.
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

> Se completa en el Sprint 0 junto con el esqueleto del proyecto (TEC-01).
