# Sprint Planning · Sprint 1 · 07/10/2026

Participantes: López, Piastrellini, Bravo · Facilitador: Juan Pablo

> La Planning correspondía al lunes 05/10 y se hizo el miércoles 07/10. Para entonces ya se habían terminado TEC-01,
> TEC-02, la tarea de godog de HU-31, HU-01, HU-30 y HU-31: se registran como parte del sprint y el recorte se hace
> sobre lo que faltaba.

## Sprint Goal

> Primer MVP: crear proyectos y sprints, registrar integrantes y cargar historias en el backlog, con los datos
> guardados en Supabase y las métricas de story points, velocidad y porcentaje de completadas ya calculadas.

## Capacidad y velocidad

- Velocidad de referencia: Sprint 0 = 21 SP (11 si se cuentan solo las tareas estimadas antes de empezar). Es un
  sprint de preparación, así que sirve como orientación, no como límite.
- Al 07/10 ya se completaron **16 SP**. Quedan 4 días (miércoles a domingo) y había **39 SP** abiertos: no entran.
- Criterio del recorte: primero lo que **no depende** de otra historia sin terminar; lo que depende de la cadena del
  backlog (HU-05 → HU-07 / HU-16 / HU-08 → HU-12 → HU-13) pasa al Sprint 2.

## Historias del sprint

### Terminadas al 07/10

| Historia | Story points | Responsable |
|---|---|---|
| TEC-01 · Esqueleto del proyecto (#37) — viene del Sprint 0 | 3 | Juan Pablo |
| TEC-02 · Integración continua (#38) — viene del Sprint 0 | 2 | Juan Pablo |
| S0 · godog con los escenarios de HU-31 (#71) — viene del Sprint 0 | 2 | Carolina |
| HU-01 · Crear proyecto (#39) | 3 | Juan Pablo |
| HU-30 · Story points planificados y completados (#8) | 3 | Carolina |
| HU-31 · Velocidad del equipo (#9) | 3 | Carolina |

### Comprometidas para el resto del sprint

| Historia | Story points | Responsable | Depende de |
|---|---|---|---|
| TEC-03 · Base de datos y migraciones (#67) | 3 | Mariano | `docs/modelo-datos.md` acordado (PR #82) |
| HU-05 · Crear ítem del Product Backlog (#18) | 3 | Mariano | — (se puede empezar con el repositorio en memoria) |
| HU-11 · Crear sprint con Sprint Goal (#43) | 3 | Juan Pablo | HU-01 (hecha) |
| HU-03 · Registrar integrantes (#41) | 3 | Juan Pablo | HU-01 (hecha) |
| HU-32 · Porcentaje de historias completadas (#10) | 2 | Carolina | — (cálculo puro, como HU-30 y HU-31) |
| HU-23 · Registrar horas estimadas (#1) | 2 | Carolina | Historia de HU-05 para guardarlas; la regla se hace antes |
| TEC-04 · Deploy en Vercel (#15) | 3 | Carolina | TEC-03: variables de Supabase en Vercel y `/health` con la base |
| Decidir cómo se conectan las métricas con el backlog (#80) | 1 | Juan Pablo y Mariano | — |

**Total del sprint:** 16 SP terminados + 20 SP comprometidos = **36 SP**.

### Pasan al Sprint 2

| Historia | Story points | Responsable | Motivo |
|---|---|---|---|
| HU-12 · Asignar historias a un sprint (#44) | 5 | Juan Pablo | Necesita historias estimadas y con criterios (HU-07, HU-16) |
| HU-13 · Registrar historias completadas (#45) | 2 | Juan Pablo | Necesita HU-12 y la regla de HU-08 |
| HU-06 · Editar y eliminar ítem (#19) | 2 | Mariano | Va después de HU-05 |
| HU-07 · Gestionar criterios de aceptación (#20) | 2 | Mariano | Va después de HU-05 |
| HU-08 · Cambiar estado de una historia (#21) | 3 | Mariano | Necesita HU-07 (solo pasa a Hecho con los criterios cumplidos) |
| HU-16 · Estimar historia con story points (#24) | 2 | Mariano | Va después de HU-05 |
| HU-24 · Registrar esfuerzo realizado (#2) | 3 | Carolina | Necesita los integrantes de HU-03 |

## Riesgos y dependencias

- **TEC-03 es el camino crítico.** Sin la base, lo que se carga en Vercel se pierde. Mientras tanto, cada historia
  sigue con el repositorio en memoria y se cambia por el de Postgres cuando esté (como HU-01).
- **Integración de las métricas (#80):** los cálculos de Carolina reciben datos simples. Se decide en este sprint
  quién los alimenta con datos reales; la conexión se programa en el Sprint 2.
- **Planning tarde:** la acción de la Retro del Sprint 0 ("ningún issue sin story points al cerrar la Planning")
  se aplica desde hoy: todas las historias del sprint tienen SP y estado *Sprint Backlog* en el tablero.

## Dailies

Lunes a viernes, por mensaje, como comentario en el issue "Sprint 1 · Dailies". Al cerrar el sprint se copian a
[`dailies.md`](dailies.md).
