# Sprint Review · Sprint 1 · 11/10/2026

Participantes: López, Piastrellini, Bravo · Product Architect (profesores): [A COMPLETAR el domingo]

> **Borrador armado el viernes 09/10** con el estado del tablero y de los PRs a esa fecha. El domingo se actualizan
> las tablas con lo que se termine el fin de semana y se completan las partes marcadas [A COMPLETAR].

## Sprint Goal

> Primer MVP: crear proyectos y sprints, registrar integrantes y cargar historias en el backlog, con los datos
> guardados en Supabase y las métricas de story points, velocidad y porcentaje de completadas ya calculadas.

¿Se cumplió? [A COMPLETAR el domingo]. Al 09/10: crear proyectos, crear sprints, registrar integrantes y las tres
métricas están **hechos**; la base en Supabase (TEC-03) está **en revisión** y cargar historias en el backlog (HU-05)
**no empezó**.

## Terminado

| Issue | Historia / tarea | SP | Responsable | Evidencia |
|---|---|---|---|---|
| #37 | TEC-01 · Esqueleto del proyecto (viene del Sprint 0) | 3 | Juan Pablo | PR #70 y arreglo del deploy en el PR #73 (Carolina) |
| #38 | TEC-02 · Integración continua (viene del Sprint 0) | 2 | Juan Pablo | PR #75 · `.github/workflows/ci.yml` |
| #71 | S0 · godog con los escenarios de HU-31 (viene del Sprint 0) | 2 | Carolina | PR #72 |
| #83 | S0 · Registro de uso de IA de Mariano | 1 | Mariano | PR #81 |
| #84 | S0 · Cerrar el modelo de datos (acordado) | 2 | Mariano | PR #82 · `docs/modelo-datos.md` |
| #39 | HU-01 · Crear proyecto | 3 | Juan Pablo | PR #78 · `/proyectos` |
| #43 | HU-11 · Crear sprint con Sprint Goal | 3 | Juan Pablo | PR #88 · `/proyectos/{id}/sprints` |
| #41 | HU-03 · Registrar integrantes | 3 | Juan Pablo | PR #90 · `/proyectos/{id}/integrantes` |
| #8 | HU-30 · Story points planificados y completados | 3 | Carolina | PR #79 · `internal/metrics/storypoints.go` |
| #9 | HU-31 · Velocidad del equipo | 3 | Carolina | PR #76 · `internal/metrics/velocity.go` |
| #10 | HU-32 · Porcentaje de historias completadas | 2 | Carolina | PR #87 · `internal/metrics/completion.go` |
| #1 | HU-23 · Registrar horas estimadas (la regla) | 2 | Carolina | PR #89 · `internal/domain/effort/horas.go` |

**Total terminado al 09/10: 29 SP** de 39 comprometidos.

## No terminado al 09/10

| Issue | Historia / tarea | SP | Responsable | Estado al 09/10 | Qué pasa |
|---|---|---|---|---|---|
| #67 | TEC-03 · Base de datos y migraciones | 3 | Mariano | En revisión (PR #91): CI en verde el 10/10 después de la revisión, falta la aprobación | [A COMPLETAR] |
| #18 | HU-05 · Crear ítem del Product Backlog | 3 | Mariano | Sin empezar | [A COMPLETAR] |
| #15 | TEC-04 · Deploy en Vercel con la base | 3 | Carolina | Esperando TEC-03 | [A COMPLETAR] |
| #80 | Decidir cómo se conectan las métricas con el backlog | 1 | Juan Pablo y Mariano | Propuesta sin confirmar | [A COMPLETAR] |

## Demo

Lo que se muestra en la app (preview o producción):

1. **Proyectos:** alta con validaciones por campo y nombre único (HU-01).
2. **Sprints de un proyecto:** número correlativo, fechas dentro del proyecto y sin superponerse (HU-11).
3. **Integrantes:** alta con rol, un solo Agile Enabler activo, email único y baja lógica (HU-03).
4. **Métricas:** se muestran con los tests y los escenarios BDD (`go test ./features/ -v`), porque la pantalla de
   métricas llega con la épica E8. [A COMPLETAR: si TEC-03 se mergea, mostrar que los datos quedan guardados]

## Feedback del cliente

[A COMPLETAR el domingo]

## Cambios en el Product Backlog

- [A COMPLETAR] Lo no terminado pasa al Sprint 2 (se decide en la Planning del lunes 12/10).
- Refinamiento de las historias del Sprint 2 en [`../sprint-2/refinamiento.md`](../sprint-2/refinamiento.md): specs de
  HU-02 y HU-04 en borrador y las preguntas abiertas de HU-12 a HU-15.
