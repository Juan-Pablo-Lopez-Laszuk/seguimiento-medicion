# Registro de uso de Inteligencia Artificial

La consigna permite usar IA siempre que el equipo comprenda, revise y valide todo lo que se incorpora.
Cada uso relevante se registra acá (el más reciente arriba).

| Fecha | Integrante | Herramienta | Para qué (historia / tarea) | Qué se usó del resultado | Cómo se validó |
|---|---|---|---|---|---|
| 06/10/2026 | Juan Pablo | Claude | HU-01 (#39): crear proyecto — spec SDD, escenarios BDD, dominio, repositorio en memoria, caso de uso y pantallas | Spec y `.feature` revisados y ajustados por mí; código y tests con ciclos RED → GREEN → REFACTOR en commits separados | Corrí cada test RED antes del código; `go test ./...` (18 escenarios BDD en verde), go vet, gofmt y golangci-lint; probé el alta y los errores por campo en la app local; revisión de Carolina en el PR |
| 05/10/2026 | Juan Pablo | Claude | TEC-02 (#38): integración continua (`.github/workflows/ci.yml`) | Workflow con gofmt, go vet, golangci-lint, tests con race y cobertura, godog y artefacto de cobertura | Simulé cada paso en local; primera corrida en verde en el PR #75; revisión de Carolina |
| 29/09/2026 | Carolina | Claude | Sprint 0: prueba "hola mundo" en Vercel (`api/index.go`, `vercel.json`, `.vercelignore`) | Los tres archivos y los comandos para publicar (`npx vercel login` y `npx vercel deploy`) | Hice yo el push y el deploy desde mi compu; abrí https://seguimiento-medicion.vercel.app y muestra "Hola mundo"; revisión de Mariano en el PR #55 |
| 28/09/2026 | Carolina | Claude | Sprint 0: plantilla BDD y escenarios de HU-31 (`features/`) | Borrador de la plantilla y de los escenarios a partir de los criterios CA-31.1 a CA-31.5 del issue #9 | Revisé que haya un escenario por criterio (normal, alternativo, límite y error) y que los resultados coincidan con `docs/metricas.md`; revisión de Mariano en el PR #36 |
| 27/09/2026 | Carolina | Claude | Sprint 0: definición y fórmula de cada métrica (`docs/metricas.md`) | Borrador del documento armado a partir de la consigna (punto 7) y de los criterios de aceptación de HU-23, HU-24 y HU-30 a HU-34 | Comparé cada fórmula con la consigna y el plan; revisión de Mariano en el PR |
| 26/09/2026 | Juan Pablo | Claude | Sprint 0: análisis de la consigna, plan de trabajo, guía de desarrollo y archivos base del repositorio | Plan, backlog inicial, plantillas, README y CONTRIBUTING | Revisión del equipo contra la consigna |
