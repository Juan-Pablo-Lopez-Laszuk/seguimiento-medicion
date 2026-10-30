# Sprint Review · Sprint 4 · 29/10/2026

Participantes: López, Piastrellini, Bravo · Product Architect (profesores): [A COMPLETAR]

> Armada el 29/10, día del congelamiento del código, a partir del tablero y de los PRs mergeados. Lo marcado
> [A COMPLETAR] se llena en la Review.

## Sprint Goal

> Cerrar el producto para la entrega: Planning Poker de punta a punta (abrir sesión, votar oculto, revelar y registrar
> la estimación), los gráficos de las métricas y la exportación del reporte a PDF; y completar la documentación, el
> informe de métricas y la presentación final.

¿Se cumplió? **Sí, en lo funcional.** La documentación y la presentación se cierran el fin de semana.
- **Planning Poker anda de punta a punta:** abrir la sesión (HU-17), votar sin ver los valores (HU-18), revelar los
  votos (HU-19) y registrar la estimación acordada (HU-22).
- **Métricas:** los gráficos (HU-36) están en la app.
- **Reportes:** el del sprint y el del proyecto se descargan en PDF (HU-39).
- **Documentación:** el informe de métricas y el manual de usuario están en `main` (#427). Faltan secciones del manual,
  la arquitectura y el guion de la presentación (ver [Pendiente para la entrega](#pendiente-para-la-entrega)).

## Terminado

### Comprometido

| Issue | Historia | SP | Responsable | Evidencia |
|---|---|---|---|---|
| #25 | HU-17 · Abrir sesión de Planning Poker | 3 | Mariano | PR #430 (modelo común de la épica) y PR #432 · `/proyectos/{id}/backlog/{n}/poker` |
| #14 | HU-36 · Gráficos de métricas | 5 | Carolina | PR #426 · `/proyectos/{id}/graficos` |
| #50 | HU-39 · Exportar reporte a PDF | 5 | Juan Pablo | PR #428 · `…/sprints/{n}/reporte.pdf` y `/proyectos/{id}/reporte.pdf` |
| #26 | HU-18 · Votar de forma oculta | 5 | Mariano | PR #433 · cartas de la escala; se ve quién votó, nunca los valores |
| #30 | HU-22 · Registrar la estimación acordada | 2 | Juan Pablo | PR #437 · la historia queda estimada y la sesión pasa a "Sesiones cerradas" |
| #27 | HU-19 · Revelar estimaciones | 3 | Carolina | PR #435 y PR #438 (arreglos de la revisión de Mariano) · se revela sola cuando votaron todos, o la cierra el Agile Enabler |

**Total: 23 SP de 23 comprometidos.** Por integrante: Juan Pablo 7 de 7 · Mariano 8 de 8 · Carolina 8 de 8.

### Adelantado (no comprometido)

| Issue | Historia | SP | Responsable | Evidencia |
|---|---|---|---|---|
| #101 | TEC-07b · Componentes de formularios y tablas | 2 | Carolina | PR #436 · el error debajo de cada campo y tablas que se apilan en el celular |

Fuera del compromiso y sin SP:
- **Seguridad en el pipeline** (#431, Juan Pablo): `govulncheck` en cada PR, en cada push a `main` y una vez por
  semana. Además, pgx, x/text y Go actualizados, lo que cerró 15 vulnerabilidades. Dependabot quedó activado y cerró su
  única alerta con #434.
- **Documentación:** el informe de métricas y el manual (#427, Carolina) y el registro de las revisiones de Mariano
  (#439).
- **Proyecto de ejemplo para la demo** (#440, Juan Pablo, en revisión).

**Velocidad del Sprint 4: 25 SP** (23 comprometidos y 2 adelantados), con 4 días de código. Sprint 1: 39 SP ·
Sprint 2: 33 SP · Sprint 3: 48 SP. **Total del proyecto: 145 SP.**

### No terminado

| Issue | Historia | SP | Qué pasó |
|---|---|---|---|
| #28 | HU-20 · Detectar diferencias entre estimaciones | 3 | No estaba comprometida ("si sobra tiempo"). Juan Pablo y Mariano la dejaron afuera el 28/10, sin empezar, antes del congelamiento. Vuelve al Product Backlog |
| #29 | HU-21 · Realizar nuevas rondas | 3 | Lo mismo que HU-20 |

Sin HU-20 la app no destaca los votos extremos. Sin HU-21, si no hay acuerdo, el Agile Enabler registra igual el valor
que se acuerde de palabra y abre otra sesión. Las sesiones cerradas quedan en el historial.

## Demo

Lo que se muestra en la app:

1. **Planning Poker** (HU-17, HU-18, HU-19 y HU-22), en una historia sin estimar:
   - el Agile Enabler abre la sesión;
   - cada integrante vota con las cartas, y la lista muestra quién votó pero no los valores;
   - cuando votaron todos, la ronda se revela sola y aparece la tabla de votos;
   - el Agile Enabler registra el valor acordado: la historia queda estimada, el cambio va al historial de estimaciones
     y la sesión queda en "Sesiones cerradas".
2. **Gráficos** (HU-36): velocidad por sprint, SP planificados contra completados, horas estimadas contra reales,
   defectos por severidad y burndown del sprint.
3. **Reportes en PDF** (HU-39): el del sprint y el del proyecto, con el mismo contenido que la pantalla.
4. **Formularios y tablas** (TEC-07b): el error debajo de cada campo y la lista de integrantes en el celular.
5. **Seguridad**: el check "Vulnerabilidades (govulncheck)" en un PR y Dependabot sin alertas.

Si falla producción o internet, el plan B es `go run ./cmd/ejemplo` (#440): la misma app en local, con un proyecto
de 3 sprints cargado.

## Calidad

- **Escenarios BDD:** 437 en verde (384 al cerrar el Sprint 3).
- **Matriz de trazabilidad:** **57 de 59** requerimientos cubiertos. Los 2 pendientes son HU-20 y HU-21.
- **CI:** exige dos checks para mergear a `main`:
  - "Calidad": gofmt, go vet, golangci-lint, los tests con PostgreSQL, los escenarios y el umbral de cobertura del 80 %;
  - "Vulnerabilidades (govulncheck)".

  La cobertura combinada que mide el CI está en 99,3 %.
- **TDD:** cada historia tiene sus commits `RED` antes de los `GREEN`, y en cada revisión se verificó commit por commit.
- **Planning Poker, verificado juntando historias:** HU-19 y HU-22 se probaron juntas en local antes de mergear. Así
  aparecieron a tiempo un nombre de error repetido y los conflictos de la página de la sesión.
- **Sin migraciones nuevas:** las tablas de Planning Poker estaban desde la 0001, así que no hubo que tocar la base de
  producción.

## Feedback del cliente

- [A COMPLETAR en la Review]

## Cambios en el Product Backlog

- **HU-20 y HU-21 quedan afuera de la entrega:** vuelven al Product Backlog sin sprint, por decisión del 28/10 (#28 y
  #29).
- **TEC-07b entró** como adelantada.
- **Planning Poker sin tiempo real:** la lista de quién votó se actualiza cada 5 segundos con HTMX, y cuando se revela
  la ronda las otras pestañas se recargan solas (#438). No hace falta WebSocket.
- **Cuentas:** el registro sigue cerrado. Las cuentas del equipo y de los profesores se crean desde Supabase con
  "Add user", con el mismo email del integrante.

## Pendiente para la entrega

Es documentación y presentación, que no se congelan con el código:

| Qué | Responsable | Estado |
|---|---|---|
| Manual de usuario: secciones 1 (ingreso), 2 (proyectos, sprints y reportes) y 3 (backlog y Planning Poker) | Juan Pablo | Listas, en el PR de documentación (#442) |
| Informe: cobertura (entregable 10) y evidencias de TDD (entregable 8) | Mariano | Viernes 30/10 |
| Informe de métricas con los datos del Sprint 4 | Carolina | [A CONFIRMAR] |
| Arquitectura (entregable 12) y guion de la presentación (entregable 13) | Juan Pablo | En el PR de documentación (#442) |
| Proyecto de ejemplo para la demo (#440) | Juan Pablo | En revisión (Carolina) |
| Ensayo de la presentación | Los tres | Sábado 31/10 o domingo 01/11 |
