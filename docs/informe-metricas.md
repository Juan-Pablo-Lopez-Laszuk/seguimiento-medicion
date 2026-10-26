# Informe de métricas del equipo

Entregable 10 de la consigna. Muestra cómo trabajó el equipo en cada sprint con las mismas métricas que calcula la
aplicación (`docs/metricas.md`): story points planificados y completados, velocidad, horas y defectos.

> Secciones 1 a 6: Carolina. Sección 7 (cobertura) y sección 8 (evidencias de TDD): Mariano.

## 1. De dónde salen los datos

| Dato | Fuente |
|---|---|
| Story points comprometidos y terminados | Las actas de Planning y Review de cada sprint (`docs/scrum/sprint-N/`) y el tablero del proyecto |
| Velocidad | Calculada con la regla de HU-31: promedio de los SP completados de los últimos 3 sprints cerrados |
| Horas estimadas | El campo "Horas estimadas" de cada tarea (sub-issues de cada historia, desde el Sprint 3) y la Planning del Sprint 4 |
| Pull requests y commits | El historial de `main` (`git log`), contando cada PR en el sprint en que se mergeó |
| Arreglos | Los commits de tipo `fix` que llegaron a `main` |

Sprints: **Sprint 0** (hasta el 04/10), **Sprint 1** (05/10 al 12/10, se estiró un día), **Sprint 2** (13/10 al 18/10),
**Sprint 3** (19/10 al 25/10) y **Sprint 4** (26/10 al 01/11, el de cierre).

## 2. Story points planificados y completados (M1 y M2)

| Sprint | Comprometidos | Terminados | Cumplimiento | Comentario |
|---|---|---|---|---|
| Sprint 0 | 18 | 21 | 116,7 % | Se sumaron tareas que se estimaron después de terminadas (marcadas con * en la Review) |
| Sprint 1 | 39 | 39 | 100,0 % | 6 SP entraron el lunes 12/10, el día que se estiró el sprint |
| Sprint 2 | 33 | 33 | 100,0 % | 6 SP entraron el último día (domingo 18/10) |
| Sprint 3 | 40 | 48 | 120,0 % | Se adelantaron HU-26 y HU-29 (5 SP) y HU-38 (3 SP) |
| **Total** | **130** | **141** | **108,5 %** | |

En todos los sprints se terminó todo lo comprometido. En el Sprint 3 se terminó más, porque las historias que destraban
a otras se hicieron primero (acción de la Retro del Sprint 1) y quedó tiempo para adelantar.

## 3. Velocidad (M3)

| Al cerrar | Sprints que promedia | Velocidad |
|---|---|---|
| Sprint 1 | 39 | 39,0 SP por sprint |
| Sprint 2 | 39 y 33 | 36,0 SP por sprint |
| Sprint 3 | 39, 33 y 48 | 40,0 SP por sprint |

El Sprint 0 no cuenta: fue de preparación (tablero, backlog, plantillas y deploy). Con 40 SP de velocidad y 4 días de
código, la Planning del Sprint 4 comprometió 23 SP (más o menos la mitad de una semana completa).

### Por integrante

| Sprint | Carolina | Juan Pablo | Mariano |
|---|---|---|---|
| Sprint 0 | 11 | 4 | 6 |
| Sprint 1 | 15 | 15 | 9 |
| Sprint 2 | 12 | 12 | 9 |
| Sprint 3 | 18 | 15 | 15 |
| **Total** | **56** | **46** | **39** |

Mariano tuvo menos SP en los Sprints 1 y 2, pero llevó las historias técnicas más pesadas (TEC-03, TEC-05 y TEC-06) y
la revisión de las historias de Carolina.

## 4. Horas estimadas y reales (M4, M5 y M6)

Desde el Sprint 3 cada historia se divide en tareas con sus **horas estimadas**, cargadas antes de empezar. En el
Sprint 4 las historias comprometidas suman **76 h** estimadas: Mariano 27,5 h (HU-17 y HU-18), Carolina 26 h (HU-36 y
HU-19) y Juan Pablo 22,5 h (HU-39 y HU-22).

Las **horas reales** del equipo no se registraron de forma sistemática: el tablero no tiene un campo para eso y las
tareas se cierran con un comentario de qué se hizo, no de cuánto llevó. Por eso este informe no calcula la desviación
del equipo. La aplicación sí la calcula para cualquier proyecto (HU-24 y HU-33): para la demostración se cargan horas
reales en el proyecto de ejemplo y se ve la desviación en la página de Horas, en el reporte y en los gráficos.

**Lección:** si se hubiera usado la propia aplicación para registrar las horas desde el Sprint 1, este punto tendría
datos reales. Queda como mejora para un próximo proyecto.

## 5. Defectos y arreglos (M8 y M9)

El equipo no cargó sus propios defectos en la aplicación: los problemas se encontraron en las revisiones de los PRs y
en el CI, y se arreglaron en el mismo PR o en uno de arreglo. Como aproximación, estos son los commits de arreglo que
llegaron a `main`:

| Sprint | Commits de arreglo | Qué arreglaron |
|---|---|---|
| Sprint 0 | 1 | TEC-01: Vercel no podía importar `internal/` (se agregó el paquete `app`) |
| Sprint 1 | 2 | TEC-03: Vercel tenía que subir `migrations/`; un mensaje de error que no cumplía el lint (ST1005) |
| Sprint 2 | 1 | HU-06: el texto de ayuda del formulario de editar |
| Sprint 3 | 2 | HU-28: un error sin revisar en un paso BDD (errcheck del CI); TEC-05: un selector que marcaba el lint (QF1008) |

Casi todos los problemas se frenaron antes de llegar a `main`: cada PR pasa por el CI (lint, tests con y sin
PostgreSQL y, desde TEC-06, el umbral de cobertura) y por la revisión de otro integrante, que prueba commit por commit
que cada RED falla y cada GREEN pasa.

## 6. Actividad del repositorio

| Sprint | PRs mergeados | Commits (sin merges) | Por integrante (commits): Carolina · Juan Pablo · Mariano |
|---|---|---|---|
| Sprint 0 | 10 | 25 | 9 · 10 · 6 |
| Sprint 1 | 17 | 241 | 68 · 114 · 59 |
| Sprint 2 | 14 | 203 | 57 · 44 · 102 |
| Sprint 3 | 20 | 196 | 61 · 58 · 77 |

En `main` hay **215 commits `RED`** y **214 commits `GREEN`**: cada historia se hizo con ciclos de TDD en commits
separados (la sección 8 explica cómo verlos).

## 7. Cobertura

> Sección a cargo de Mariano: cobertura por paquete, el umbral de TEC-06 y cómo se mide.

## 8. Evidencias de TDD

> Sección a cargo de Mariano: cómo ver los commits `RED` → `GREEN` de cada historia, con dos o tres ejemplos.
