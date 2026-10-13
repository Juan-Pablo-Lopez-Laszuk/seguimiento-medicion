# SPEC-04 · Consultar estado del proyecto

**Historia:** HU-04 (#42) · **Responsable:** López, Juan Pablo · **Sprint:** 2 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`, las tablas Proyecto, Sprint e
> Historia de `docs/modelo-datos.md` y las métricas M1 y M2 de `docs/metricas.md`.

## 1. Objetivo

Que cualquier integrante vea en una sola pantalla cómo va el proyecto: en qué estado está, cuál es el sprint activo,
cuántos días quedan y qué porcentaje de los story points ya se completó. Además, esta pantalla pasa a ser la **vista
del proyecto**, desde donde se entra a sus sprints y a sus integrantes.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Hoy | fecha | — | No la carga el usuario: es la fecha actual **en la hora de Argentina** (RN6) |

## 3. Salidas esperadas

La página `/proyectos/{id}` muestra:

- Nombre, descripción y fechas del proyecto.
- **Estado** del proyecto: Planificado, En curso o Finalizado (RN1).
- **Sprint activo** (RN2): número, Sprint Goal, fechas y días que le quedan. Si no hay, el texto "Sin sprint activo"
  y, si existe, cuál es el próximo y cuándo empieza.
- **Días restantes del proyecto** (RN3).
- **Porcentaje de story points completados** del proyecto (RN4), con 1 decimal y coma (por ejemplo "37,5 %").
- Enlaces a los **Sprints** y a los **Integrantes** del proyecto. En la lista de proyectos, el nombre pasa a ser el
  enlace a esta vista.

## 4. Reglas de negocio

- RN1: **Estado del proyecto.** No se guarda: se calcula cada vez a partir de las fechas y los sprints (CA-04.2).
  - Sin sprints → **Planificado**, sean cuales sean las fechas (CA-04.3).
  - Con sprints y hoy **antes** del inicio del primer sprint → **Planificado**.
  - Con sprints y hoy **después** de la fecha de fin del proyecto → **Finalizado**.
  - En cualquier otro caso → **En curso**.
- RN2: **Sprint activo.** Es el sprint **no cerrado** de **menor número** cuya fecha de inicio ya llegó (inicio ≤ hoy).
  Como los sprints se crean en orden (HU-11 RN4), hay como máximo uno. Un sprint cuya fecha de fin ya pasó sigue activo
  hasta que se cierra (HU-14): así se puede cerrar aunque se haya cerrado tarde. Esta regla la usan también HU-13
  ("solo historias del sprint activo") y HU-14 ("solo se cierra el sprint activo"), por eso vive en
  `internal/domain/sprint`.
- RN3: **Días restantes** = días corridos desde hoy hasta la fecha de fin, sin contar hoy. El último día da 0 y una
  fecha ya pasada también da 0 (nunca negativo). Se calcula para el proyecto y para el sprint activo; si el sprint activo
  ya terminó por fecha, se muestra "Terminó el dd/mm/aaaa: falta cerrarlo".
- RN4: **Porcentaje de SP completados** = SP de las historias Hechas ÷ SP de todas las historias estimadas del
  proyecto × 100. Se calcula con `metrics.CalcularStoryPoints` (M1 y M2, HU-30). Las historias sin estimar no suman y se
  avisa cuántas son. Si no hay SP estimados, se muestra "Sin historias estimadas" en lugar de 0 % (RG-3: nunca se
  divide por cero). Hasta que lleguen HU-08 y HU-13 todas las historias nacen y quedan en "Pendiente" (HU-05 RN6), así
  que el porcentaje va a dar 0 %: es lo esperado, no un error.
- RN5: La pantalla **solo lee**: no cambia nada del proyecto, de los sprints ni de las historias.
- RN6: "Hoy" se toma en la **zona horaria de Argentina** (UTC−3). Vercel corre en UTC: entre las 21:00 y las 24:00 de
  Argentina, en UTC ya es el día siguiente y el estado o los días restantes darían mal.

## 5. Restricciones

- RN1, RN2 y RN3 son **funciones puras** en el dominio (`project.EstadoActual`, `sprint.Activo` y `domain.DiasRestantes`)
  que reciben "hoy" como parámetro, así los tests no dependen del reloj.
- El caso de uso arma la vista con el proyecto, sus sprints y sus historias. Las historias se leen con
  `RepositorioHistorias.ListarPorProyecto` de HU-05 (ya en `main`) y se pasan a `metrics.Historia` (`Hecha` = estado
  Hecho, `SinEstimar` = sin story points).
- La columna `proyectos.estado` de la migración de TEC-03 queda con el valor con el que se creó: el estado que se
  muestra es siempre el calculado (RN1). Se puede quitar más adelante, sin apuro.

## 6. Casos límite

Proyecto de ejemplo: del 05/10/2026 al 01/11/2026. Sprint 1 del 05/10 al 11/10 y Sprint 2 del 12/10 al 18/10.

| Caso | Resultado |
|---|---|
| Proyecto sin sprints, hoy 20/10 | Planificado (CA-04.3) |
| Proyecto sin sprints, hoy 10/11 (ya pasó el fin) | Planificado (CA-04.3) |
| Con sprints, hoy 04/10 (antes del primer sprint) | Planificado, sin sprint activo, próximo: Sprint 1 el 05/10 |
| Con sprints, hoy 05/10 (empieza el Sprint 1) | En curso, Sprint 1 activo, le quedan 6 días |
| Hoy 11/10 (último día del Sprint 1) | En curso, Sprint 1 activo, le quedan 0 días |
| Hoy 13/10 y el Sprint 1 no está cerrado | Sprint 1 sigue activo: "Terminó el 11/10/2026: falta cerrarlo" |
| Hoy 13/10 y el Sprint 1 está cerrado | Sprint 2 activo, le quedan 5 días |
| Hoy 01/11 (último día del proyecto) | En curso, le quedan 0 días |
| Hoy 02/11 | Finalizado, le quedan 0 días |
| Historias Hechas con 3 SP de 8 estimados | 37,5 % |
| Sin historias o todas sin estimar | "Sin historias estimadas" |
| A las 22:00 de Argentina del 01/11 (en UTC ya es 02/11) | En curso (RN6) |
| El proyecto no existe | Página "no encontrado" (404) |

## 7. Condiciones de error

| Situación | Error (Go) | Mensaje al usuario |
|---|---|---|
| El proyecto no existe | `project.ErrNoEncontrado` | Página 404 "No encontramos ese proyecto" |
| Falla la base | error del repositorio | "No se pudo cargar el proyecto. Probá de nuevo en unos minutos." (500, sin detalles internos) |

## 8. Criterios de aceptación

- CA-04.1: Muestra estado (Planificado / En curso / Finalizado), sprint activo, días restantes y % de SP completados.
- CA-04.2: El estado se deriva de las fechas y los sprints según la regla de la spec.
- CA-04.3: Un proyecto sin sprints se muestra como Planificado.

## Trazabilidad

Issue: #42 · Feature: `features/hu-04-estado-proyecto.feature` · Pasos BDD: `features/hu04_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Dominio | `internal/domain/project` (`EstadoActual`), `internal/domain/sprint` (`Activo`), `internal/domain` (`DiasRestantes`) | `*_test.go` de cada paquete |
| Caso de uso | `internal/service/proyectos.go` (`Estado`) | `internal/service/proyectos_test.go` |
| Pantalla | `internal/server/proyectos.go`, `web/templates/paginas/proyecto.html` | `internal/server/proyectos_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-04.1 Muestra estado, sprint activo, días y % de SP | Proyecto en curso con sprint activo · Sprint terminado sin cerrar · Porcentaje de SP | se completa con el código |
| CA-04.2 Estado derivado de fechas y sprints | Estado según la fecha | se completa con el código |
| CA-04.3 Sin sprints es Planificado | Proyecto sin sprints | se completa con el código |
