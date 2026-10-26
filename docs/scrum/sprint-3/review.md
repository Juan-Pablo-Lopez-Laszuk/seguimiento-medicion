# Sprint Review · Sprint 3 · 25/10/2026

Participantes: López, Piastrellini, Bravo · Product Architect (profesores): [A COMPLETAR]

> Borrador armado el 24/10 a partir del tablero y de los PRs mergeados. Lo marcado [A COMPLETAR] se llena en la Review.

## Sprint Goal

> Métricas completas y a la vista: cerrar sprints con su foto, ver los sprints anteriores, las horas y los defectos
> por sprint, el panel de indicadores del proyecto y el reporte de un sprint; y sostener la calidad con pruebas de
> integración y un umbral de cobertura en el CI.

¿Se cumplió? **Sí.**
- Se cierra el sprint con su foto (HU-14) y se ven los sprints anteriores (HU-15).
- Están las horas y la desviación (HU-33) y los defectos detectados y resueltos (HU-34).
- Están el panel del proyecto (HU-35) y el reporte de un sprint (HU-37).
- El CI exige un 80 % de cobertura y la matriz de trazabilidad tiene un test que la mantiene al día (TEC-06).

Además, la app tiene login en producción (TEC-05, adelantada del Sprint 4).

## Terminado

### Comprometido

| Issue | Historia / tarea | SP | Responsable | Evidencia |
|---|---|---|---|---|
| #46 | HU-14 · Cerrar sprint | 5 | Juan Pablo | PR #117 · `/proyectos/{id}/sprints/{n}/cerrar` |
| #6 | HU-28 · Gestionar el estado de un defecto | 3 | Carolina | PR #112 · estado en la página de defectos |
| #11 | HU-33 · Horas estimadas, reales y desviación | 3 | Carolina | PR #116 · `/proyectos/{id}/horas` |
| #106 | HU-16b · Historial de estimaciones | 2 | Mariano | PR #414 · historial en la página de la historia (migración 0003, aplicada en producción el 21/10) |
| #69 | TEC-06 · Pruebas de integración y umbral de cobertura | 3 | Mariano | PR #415 (umbral del 80 % en el CI) y PR #419 (matriz de trazabilidad con su test) |
| #47 | HU-15 · Consultar sprints anteriores | 2 | Juan Pablo | PR #413 · SP completados y cumplimiento en la lista de sprints, y la foto en cada sprint cerrado |
| #12 | HU-34 · Defectos detectados y resueltos | 2 | Carolina | PR #119 · defectos por sprint |
| #22 | HU-09 · Priorizar el backlog | 3 | Mariano | PR #421 · Subir y Bajar dentro de cada prioridad |
| #68 | TEC-05 · Autenticación de usuarios | 5 | Mariano | PR #418 · `/login` y `/registro`; activo en producción desde el 23/10 |
| #48 | HU-37 · Reporte de sprint | 5 | Juan Pablo | PR #416 · `/proyectos/{id}/sprints/{n}/reporte` |
| #13 | HU-35 · Panel de indicadores del proyecto | 5 | Carolina | PR #120 · `/proyectos/{id}/panel` |
| #23 | HU-10 · Filtrar y buscar historias | 2 | Mariano | PR #422 · filtros y búsqueda en el backlog |

**Total: 40 SP de 40 comprometidos.** Por integrante: Juan Pablo 12 de 12 · Mariano 15 de 15 · Carolina 13 de 13.

### Adelantado (no comprometido)

| Issue | Historia | SP | Responsable | Evidencia |
|---|---|---|---|---|
| #7 | HU-29 · Consultar defectos | 2 | Carolina | PR #412 · filtros en la página de defectos |
| #4 | HU-26 · Consultar y corregir registros de esfuerzo | 3 | Carolina | PR #417 · `/proyectos/{id}/esfuerzo` |
| #49 | HU-38 · Reporte de proyecto | 3 | Juan Pablo | PR #423 (mergeado el 25/10) · `/proyectos/{id}/reporte` |

Fuera del compromiso y sin SP: arreglo de HU-06 en el modo memoria (PR #420, Mariano) y la Planning del Sprint 4 (#425).

**Velocidad del Sprint 3: 48 SP** (40 comprometidos y 8 adelantados). Sprint 1: 39 SP · Sprint 2: 33 SP.

## Demo

Lo que se muestra en la app:

1. **Login** (TEC-05): registrarse, ingresar y salir. Sin sesión, todas las páginas piden ingresar.
2. **Cerrar un sprint** (HU-14):
   - la confirmación avisa si faltan días;
   - las historias sin terminar vuelven al backlog;
   - el sprint queda con su foto.
3. **Sprints anteriores** (HU-15): SP completados y cumplimiento de cada sprint cerrado. Si después se re-estima una
   historia (HU-16b, con su historial), la foto no cambia.
4. **Backlog** (HU-09 y HU-10): priorizar con Subir y Bajar, y filtrar o buscar historias.
5. **Horas** (HU-33 y HU-26): estimadas, reales y desviación por historia y por sprint; consultar y corregir los
   registros.
6. **Defectos** (HU-28, HU-29 y HU-34): cambiar el estado, filtrar y ver los detectados y resueltos por sprint.
7. **Panel del proyecto** (HU-35): velocidad, avance, horas y defectos en una sola página, con "sin datos" donde todavía
   no hay.
8. **Reportes** (HU-37 y HU-38): el reporte de un sprint y el del proyecto, con la evolución de la velocidad. Se
   imprimen con cada sprint en una hoja nueva.

## Calidad

- 384 escenarios BDD en verde y la matriz de trazabilidad en **51 de 59** requerimientos (52 con HU-39, en el Sprint 4).
- El CI corre gofmt, go vet, golangci-lint, los tests con PostgreSQL, los escenarios y el umbral de cobertura del 80 %.
  Hoy `internal/domain` e `internal/metrics` están entre el 97 % y el 100 %.
- Cada historia tiene sus commits `RED` antes de los `GREEN`. En cada revisión se verificó commit por commit.

## Feedback del cliente

- [A COMPLETAR en la Review]

## Cambios en el Product Backlog

- **TEC-05 se adelantó** del Sprint 4 (propuesta de Mariano, 19/10). Planning Poker la necesita para saber quién vota.
- **Registro de cuentas cerrado** (decisión de Juan Pablo, 24/10): solo el equipo tiene cuenta. Cualquier cuenta ve
  todos los proyectos; los permisos por rol quedan fuera del alcance.
- **Pasan al Sprint 4:** Planning Poker (HU-17 a HU-22) y HU-36 (gráficos), como estaba previsto, más TEC-07b (#101).
  El reparto está en la [Planning del Sprint 4](../sprint-4/planning.md).
