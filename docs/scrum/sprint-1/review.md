# Sprint Review · Sprint 1 · 11/10/2026

Participantes: López, Piastrellini, Bravo · Product Architect (profesores): no estuvieron

## Sprint Goal

> Primer MVP: crear proyectos y sprints, registrar integrantes y cargar historias en el backlog, con los datos
> guardados en Supabase y las métricas de story points, velocidad y porcentaje de completadas ya calculadas.

¿Se cumplió? **Parcialmente.** Al cierre (domingo 11/10) estaban hechos crear proyectos, crear sprints, registrar
integrantes, las tres métricas y la base de datos en Supabase (TEC-03). Faltaban cargar historias en el backlog (HU-05,
en revisión) y conectar la base en producción (TEC-04: el código estaba mergeado, faltaba cargar la base en Vercel).
Las dos se terminaron el lunes 12/10, antes de la Planning del Sprint 2.

Las tablas muestran el estado **al cierre del sprint, el domingo 11/10**. Lo que se terminó después cuenta para el
Sprint 2, igual que en el Sprint 0.

## Terminado al cierre (11/10)

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
| #67 | TEC-03 · Base de datos y migraciones | 3 | Mariano | PR #91 · `migrations/`, `internal/store/postgres` |
| #80 | Decidir cómo se conectan las métricas con el backlog | 1 | Juan Pablo y Mariano | Decisión publicada en el issue (10/10) |

**Total terminado al cierre: 33 SP** de 39 comprometidos.

## No terminado al cierre (pasa al Sprint 2)

| Issue | Historia / tarea | SP | Responsable | Estado al 11/10 | Qué pasó después |
|---|---|---|---|---|---|
| #18 | HU-05 · Crear ítem del Product Backlog | 3 | Mariano | En revisión (PR #93, abierto el sábado 10/10 a la noche) | Aprobado y mergeado el lunes 12/10 (PR #93) · `/proyectos/{id}/backlog` |
| #15 | TEC-04 · Deploy en Vercel con la base | 3 | Carolina | Código mergeado (PR #92); faltaba cargar la base de producción en Vercel | El lunes 12/10 Juan Pablo aplicó las migraciones en producción y cargó `DATABASE_URL` en Vercel; `/health` responde `"base":"ok"` y Carolina cerró el issue |

Como las dos se terminaron después del cierre, sus 6 SP cuentan para el Sprint 2 ("viene del Sprint 1").

## Demo

Lo que se muestra en la app:

1. **Proyectos:** alta con validaciones por campo y nombre único (HU-01).
2. **Sprints de un proyecto:** número correlativo, fechas dentro del proyecto y sin superponerse (HU-11).
3. **Integrantes:** alta con rol, un solo Agile Enabler activo, email único y baja lógica (HU-03).
4. **Datos guardados en Supabase** (TEC-03): con `DATABASE_URL` cargada, lo que se crea queda guardado al reiniciar la
   app. Al cierre funcionaba con la base de desarrollo; en producción, desde el 12/10 (TEC-04).
5. **Métricas:** se muestran con los tests y los escenarios BDD (`go test ./features/ -v`), porque la pantalla de
   métricas llega con la épica E8.

## Feedback del cliente

- No hubo feedback durante la Review porque los profesores no estuvieron.

## Cambios en el Product Backlog

- HU-05 (#18) y TEC-04 (#15) pasan al Sprint 2 como terminadas ("viene del Sprint 1"); en el tablero quedan asignadas
  al Sprint 2, como se hizo con TEC-01 y TEC-02 en el Sprint 0.
- Refinamiento de las historias del Sprint 2 en [`../sprint-2/refinamiento.md`](../sprint-2/refinamiento.md): specs de
  HU-02 y HU-04 en borrador y las preguntas abiertas de HU-12 a HU-15. HU-05 ya está en `main`, así que la cadena de
  dependencias del backlog (HU-07, HU-08, HU-16, HU-25) queda desbloqueada para el Sprint 2.
