# Sprint Review · Sprint 1 · 12/10/2026

Participantes: López, Piastrellini, Bravo · Product Architect (profesores): no estuvieron

## Sprint Goal

> Primer MVP: crear proyectos y sprints, registrar integrantes y cargar historias en el backlog, con los datos
> guardados en Supabase y las métricas de story points, velocidad y porcentaje de completadas ya calculadas.

¿Se cumplió? **Sí**, con un día de atraso. El domingo 11/10 faltaban cargar historias en el backlog (HU-05, en
revisión) y conectar la base en producción (TEC-04: el código estaba mergeado, faltaba cargar la base en Vercel). Como
el Sprint 2 todavía no había empezado, el equipo decidió **estirar el Sprint 1 hasta el lunes 12/10** para cerrarlas, y
las dos cuentan para este sprint.

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
| #67 | TEC-03 · Base de datos y migraciones | 3 | Mariano | PR #91 · `migrations/`, `internal/store/postgres` |
| #80 | Decidir cómo se conectan las métricas con el backlog | 1 | Juan Pablo y Mariano | Decisión publicada en el issue (10/10) |
| #18 | HU-05 · Crear ítem del Product Backlog (12/10) | 3 | Mariano | PR #93 · `/proyectos/{id}/backlog` |
| #15 | TEC-04 · Deploy en Vercel con la base (12/10) | 3 | Carolina | PR #92 · `/health` responde `"base":"ok"` en producción (migraciones aplicadas y `DATABASE_URL` cargada en Vercel por Juan Pablo) |

**Total terminado: 39 SP** de 39 comprometidos (33 al domingo 11/10 y 6 el lunes 12/10).

## Demo

Lo que se muestra en la app:

1. **Proyectos:** alta con validaciones por campo y nombre único (HU-01).
2. **Sprints de un proyecto:** número correlativo, fechas dentro del proyecto y sin superponerse (HU-11).
3. **Integrantes:** alta con rol, un solo Agile Enabler activo, email único y baja lógica (HU-03).
4. **Product Backlog:** alta de historias con título, prioridad y story points, con validaciones por campo (HU-05).
   Los criterios de aceptación llegan con HU-07.
5. **Datos guardados en Supabase** (TEC-03 y TEC-04): en producción, lo que se crea queda guardado al recargar y al
   redesplegar.
6. **Métricas:** se muestran con los tests y los escenarios BDD (`go test ./features/ -v`), porque la pantalla de
   métricas llega con la épica E8.

## Feedback del cliente

- No hubo feedback durante la Review porque los profesores no estuvieron.

## Cambios en el Product Backlog

- No queda nada pendiente del Sprint 1: HU-05 (#18) y TEC-04 (#15) quedan en el Sprint 1 en el tablero.
- Refinamiento de las historias del Sprint 2 en [`../sprint-2/refinamiento.md`](../sprint-2/refinamiento.md): specs de
  HU-02 y HU-04 en borrador y las preguntas abiertas de HU-12 a HU-15. HU-05 ya está en `main`, así que la cadena de
  dependencias del backlog (HU-07, HU-08, HU-16, HU-25) queda desbloqueada para el Sprint 2.
