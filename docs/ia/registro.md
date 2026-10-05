# Registro de uso de Inteligencia Artificial

La consigna permite usar IA siempre que el equipo comprenda, revise y valide todo lo que se incorpora.
Cada uso relevante se registra acá (el más reciente arriba).

| Fecha | Integrante | Herramienta | Para qué (historia / tarea) | Qué se usó del resultado | Cómo se validó |
|---|---|---|---|---|---|
| 05/10/2026 | Juan Pablo | Claude | Sprint 0: borradores de las actas de Review y Retro (`docs/scrum/sprint-0/`) | Estructura de las actas, historias y SP tomados del tablero, propuestas para las preguntas abiertas | Las opiniones y acciones de la Retro las completa el equipo; las respuestas de los profesores se cargan a mano cuando lleguen |
| 01/10/2026 | Juan Pablo | Claude | TEC-01 (#37): esqueleto del proyecto (`go.mod`, `internal/server`, `web/templates`, `cmd/server`, `api/index.go`, runner de godog) | Código y tests de `/health`, layout base, página de inicio y 404, con los ciclos RED → GREEN en commits separados | `go test ./...`, `go vet` y `gofmt`; levanté la app en local y revisé `/`, `/health` y una ruta inexistente; el preview de Vercel falló por importar `internal/` y Carolina lo corrigió en el PR #73 |
| 29/09/2026 | Carolina | Claude | Sprint 0: prueba "hola mundo" en Vercel (`api/index.go`, `vercel.json`, `.vercelignore`) | Los tres archivos y los comandos para publicar (`npx vercel login` y `npx vercel deploy`) | Hice yo el push y el deploy desde mi compu; abrí https://seguimiento-medicion.vercel.app y muestra "Hola mundo"; revisión de Mariano en el PR #55 |
| 28/09/2026 | Carolina | Claude | Sprint 0: plantilla BDD y escenarios de HU-31 (`features/`) | Borrador de la plantilla y de los escenarios a partir de los criterios CA-31.1 a CA-31.5 del issue #9 | Revisé que haya un escenario por criterio (normal, alternativo, límite y error) y que los resultados coincidan con `docs/metricas.md`; revisión de Mariano en el PR #36 |
| 27/09/2026 | Carolina | Claude | Sprint 0: definición y fórmula de cada métrica (`docs/metricas.md`) | Borrador del documento armado a partir de la consigna (punto 7) y de los criterios de aceptación de HU-23, HU-24 y HU-30 a HU-34 | Comparé cada fórmula con la consigna y el plan; revisión de Mariano en el PR |
| 26/09/2026 | Juan Pablo | Claude | Sprint 0: análisis de la consigna, plan de trabajo, guía de desarrollo y archivos base del repositorio | Plan, backlog inicial, plantillas, README y CONTRIBUTING | Revisión del equipo contra la consigna |
