# Software Metrics & Estimation

Aplicación web para **administrar proyectos de software y medir su estimación, planificación, seguimiento y calidad**.
Permite gestionar el Product Backlog y los Sprints, estimar historias con **Planning Poker**, registrar esfuerzo y defectos,
y consultar **métricas, un dashboard y reportes** (exportables a PDF).

Trabajo Práctico Integrador · **Ingeniería y Calidad de Software 2026** · UTN Facultad Regional San Rafael.

- **Tablero Scrum (GitHub Projects):** https://github.com/users/Juan-Pablo-Lopez-Laszuk/projects/1
- **Consigna, plan de trabajo y guía:** [`docs/plan/`](docs/plan/)

## Equipo

| Apellido y Nombre | GitHub | Rol Scrum | Responsable de |
|---|---|---|---|
| López, Juan Pablo | [@Juan-Pablo-Lopez-Laszuk](https://github.com/Juan-Pablo-Lopez-Laszuk) | Agile Enabler · Product Builder | Proyectos, Sprints, Reportes, arquitectura y CI |
| Piastrellini, Mariano | [@MarianoPiastre](https://github.com/MarianoPiastre) | Product Builder | Product Backlog, Planning Poker, base de datos y login |
| Bravo, Carolina | [@CaritoNBravo](https://github.com/CaritoNBravo) | Product Builder | Esfuerzo, Defectos, Métricas, Dashboard, deploy y diseño |

**Product Architect (Cliente):** profesores de la cátedra.

## Funcionalidades

Proyectos e integrantes · Product Backlog · Sprints · Estimación con Planning Poker · Registro de esfuerzo ·
Defectos · Métricas (velocidad, SP, horas, desviación, defectos) · Dashboard con gráficos · Reportes en PDF.

## Cómo lo construimos

- **Scrum** con sprints de una semana; actas de cada ceremonia en [`docs/scrum/`](docs/scrum/).
- **SDD → BDD → TDD:** cada funcionalidad tiene su especificación ([`specs/`](specs/)), escenarios Given–When–Then en español
  ([`features/`](features/)) y tests escritos antes del código (ciclo RED → GREEN → REFACTOR visible en los commits).
- **Trazabilidad:** Historia de Usuario → Especificación → Criterios de Aceptación → Escenarios BDD → Tests → Código Go.
- **IA** como apoyo, con registro de uso en [`docs/ia/registro.md`](docs/ia/registro.md).

**Stack:** Go · HTMX + Bootstrap 5 + Chart.js · Supabase (PostgreSQL + Auth) · Vercel · godog.

## Sprints

| Sprint | Fechas | Objetivo |
|---|---|---|
| 0 · Preparación | 28/09 – 04/10 | Entorno, tablero y Product Backlog inicial |
| 1 · MVP | 05/10 – 11/10 | Versión funcional básica |
| 2 · Interfaz | 12/10 – 18/10 | Interfaz usable, login y Planning Poker |
| 3 · Funcionalidad y Calidad | 19/10 – 25/10 | Dashboard, métricas completas y reportes |
| 4 · Cierre | 26/10 – 01/11 | Exportación a PDF, documentación y entrega final |

## Cómo contribuir

Ramas, convención de commits y Pull Requests: ver [CONTRIBUTING.md](CONTRIBUTING.md).
Las instrucciones para ejecutar la aplicación se agregan con el esqueleto del proyecto (Sprint 0).
