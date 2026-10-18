# SPEC-12 · Asignar historias a un sprint

**Historia:** HU-12 (#44) · **Responsable:** López, Juan Pablo · **Sprint:** 2 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa las tablas Historia, CriterioAceptacion y Sprint de
> `docs/modelo-datos.md`, el sprint activo de `specs/HU-04-estado-proyecto.md` (RN2) y la velocidad de
> `specs/HU-31-velocidad.md`. Usa HU-07 (criterios) y HU-08 (estados), de Mariano: las reglas
> que vienen de esas historias están marcadas.

## 1. Objetivo

Que el equipo arme el Sprint Backlog: elegir del Product Backlog las historias listas y comprometerlas en un sprint,
con un aviso si se compromete más de lo que el equipo suele terminar.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Sprint | número del sprint (viene en la URL) | Sí | Tiene que existir en el proyecto y no estar cerrado (RN2) |
| Historia | id de la historia (viene en el formulario) | Sí | Tiene que ser del proyecto y estar lista (RN1) |

## 3. Salidas esperadas

- Una página por sprint, **`/proyectos/{id}/sprints/{numero}`**, con el Sprint Goal, las fechas, las **historias
  asignadas** (identificador, título, SP y estado) y el **total de SP comprometidos**.
- Debajo, las **historias candidatas**: las del backlog que están en Pendiente y no tienen sprint. Cada una tiene el
  botón **Asignar** si está lista, o el motivo por el que todavía no se puede ("sin estimar", "sin criterios de
  aceptación").
- Al asignar, la historia pasa a **En sprint** (CA-08.1), queda en ese sprint y se vuelve a la página del sprint con
  el aviso "HU-N asignada al Sprint M".
- Si con esa historia los SP comprometidos superan la velocidad del equipo, el aviso lo dice (RN5). No bloquea.
- En la lista de sprints, el número de cada sprint pasa a ser el enlace a esta página.

## 4. Reglas de negocio

- RN1: **Historia lista** (CA-12.1). Solo se asigna una historia que:
  - está **estimada** (tiene story points; 0 cuenta como estimada, "sin estimar" no — igual que HU-05),
  - tiene **al menos un criterio de aceptación** (CA-07.1, viene de HU-07), y
  - está en **Pendiente** y sin sprint. Esto cubre "no Hecha" y "en un solo sprint abierto" (CA-12.2): una historia
    En sprint o En progreso ya está en un sprint abierto, y una Hecha ya terminó.
- RN2: **Sprint abierto** (CA-12.2). No se asigna a un sprint **cerrado**. Sí al sprint activo y a uno futuro que
  todavía no empezó, para poder planificar el siguiente antes de que arranque. Confirmada por Juan Pablo (13/10).
- RN3: **Una historia, un sprint abierto** (CA-12.2). Si la historia ya está en otro sprint abierto se rechaza con el
  número de ese sprint. Para cambiarla de sprint primero se la quita (RN4).
- RN4: **Quitar del sprint.** Mientras está **En sprint** (todavía no empezó), una historia se puede quitar: vuelve a
  Pendiente y sin sprint. Ya En progreso no se quita. No lo pide la historia, pero sin esto un error al planificar no
  tiene vuelta atrás. Confirmada por Juan Pablo (13/10).
- RN5: **Aviso por velocidad** (CA-12.3). Después de asignar, si los SP comprometidos del sprint superan la velocidad
  del equipo, se avisa: "Ojo: el Sprint 2 suma 21 SP y la velocidad del equipo es 16,5 SP". La velocidad es la de
  HU-31 (`metrics.CalcularVelocidad`, últimos 3 sprints cerrados, con su foto del cierre). Con un solo sprint cerrado,
  es la de ese sprint; **sin sprints cerrados no hay aviso** (decisión 8 de la Planning, confirmada por Juan Pablo y Carolina en el #96).
- RN6: El cambio de estado (Pendiente → En sprint, y la vuelta al quitar) usa las **transiciones de HU-08**, no las
  repite (decisión 5 de la Planning, confirmada por Mariano en el #95): `AsignarA` y `QuitarDelSprint` llaman a
  `Historia.CambiarEstado`.

## 5. Restricciones

- RN1, RN3 y RN4 son funciones puras en el dominio que reciben la historia, la cantidad de criterios y el sprint.
- La historia guarda en qué sprint abierto está (`historias.sprint_id`, ya en la migración 0001). Hace falta que
  `backlog.Historia` tenga el campo y que el repositorio de historias pueda **actualizar el estado y el sprint**.
  HU-08 guarda solo el estado (`ActualizarEstado`); HU-12 suma **`ActualizarSprint`**, que guarda el estado y el
  sprint **en una sola sentencia**, así una historia nunca queda En sprint sin sprint si algo falla a mitad de camino.
- Si una historia tiene criterios lo dice `backlog.ProgresoDe(criterios).TieneCriterios()` de HU-07 (ya en `main`), y
  para la lista de candidatas se usa `service.Criterios.Progresos(ctx, historias)`, que trae el progreso de todas con
  una sola consulta.
- No hay migraciones nuevas: `historias.sprint_id` y `historias.estado` ya existen.
- La velocidad lee los SP completados de los sprints cerrados (`sprints.sp_completados`, la foto que guarda HU-14,
  decisión del #80). Hasta que exista HU-14 no hay sprints cerrados, así que el aviso nunca aparece: es lo esperado.
- HU-12 se programó con HU-07 y HU-08 ya en `main`. No necesita HU-16: los story points se cargan al crear la
  historia (HU-05); HU-16 suma cambiarlos después.

## 6. Casos límite

Proyecto "Software Metrics" con el Sprint 1 (cerrado, 16 SP completados) y el Sprint 2 (activo). Historias del backlog:

| Historia | SP | Criterios | Estado |
|---|---|---|---|
| HU-1 · Login | 5 | 2 | Pendiente |
| HU-2 · Registro | sin estimar | 1 | Pendiente |
| HU-3 · Perfil | 3 | 0 | Pendiente |
| HU-4 · Reportes | 0 | 1 | Pendiente |
| HU-5 · Ayuda | 2 | 1 | Hecho |

| Caso | Resultado |
|---|---|
| Asignar HU-1 al Sprint 2 | Queda En sprint en el Sprint 2 |
| Asignar HU-4 (0 SP) | Se asigna: 0 es una estimación |
| Asignar HU-2 | Error: no está estimada |
| Asignar HU-3 | Error: no tiene criterios de aceptación |
| Asignar HU-5 | Error: ya está Hecha |
| Asignar HU-1 al Sprint 2 y después al Sprint 3 | Error: ya está en el Sprint 2 |
| Asignar al Sprint 1 (cerrado) | Error: el sprint está cerrado |
| Asignar al Sprint 3 (futuro, todavía no empezó) | Se asigna (RN2) |
| Con HU-1 (5 SP) el Sprint 2 suma 18 SP y la velocidad es 16 | Se asigna, con el aviso de velocidad |
| Sin sprints cerrados | Se asigna, sin aviso |
| Quitar HU-1 (En sprint) | Vuelve a Pendiente y sin sprint |
| Quitar una historia En progreso | Error: ya empezó |
| La historia es de otro proyecto | Página "no encontrado" (404) |

## 7. Condiciones de error

| Situación | Error (Go) | Mensaje al usuario |
|---|---|---|
| Sin estimar | `ErrSinEstimar` | "HU-2 no se puede asignar: no está estimada" |
| Sin criterios | `ErrSinCriterios` | "HU-3 no se puede asignar: no tiene criterios de aceptación" |
| Hecha | `ErrYaHecha` | "HU-5 no se puede asignar: ya está Hecha" |
| En otro sprint abierto | `ErrEnOtroSprint` | "HU-1 ya está en el Sprint 2" |
| Sprint cerrado | `ErrSprintCerrado` | "el Sprint 1 está cerrado: no se le pueden asignar historias" |
| Quitar una historia que ya empezó | `ErrYaEmpezada` | "HU-1 ya está En progreso: no se puede quitar del sprint" |
| Proyecto, sprint o historia que no existen | `project.ErrNoEncontrado` y equivalentes | Página 404 |

Los errores se muestran como aviso en la página del sprint (409), no como error de un campo: el formulario es solo el
botón Asignar.

## 8. Criterios de aceptación

- CA-12.1: Sólo historias estimadas, con CA y no Hechas.
- CA-12.2: No se asignan historias a un sprint cerrado; una historia está en un solo sprint abierto.
- CA-12.3: Advertencia (no bloqueo) si los SP comprometidos superan la velocidad del equipo (HU-31).

## Trazabilidad

Issue: #44 · Feature: `features/hu-12-asignar-historias.feature` · Pasos BDD: `features/hu12_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Dominio | `internal/domain/backlog/asignar.go` (`AsignarA`, `QuitarDelSprint`, `DatosSprint`), `Historia.SprintID`, `sprint.Sprint.SPCompletados` y `sprint.ErrNoEncontrado` | `backlog/asignar_test.go` |
| Repositorios | `ActualizarSprint` en `internal/store/{memory,postgres}/historias.go`; `sprint_id` al leer historias y `sp_completados` en `postgres/sprints.go` | `memory/historias_test.go`, `postgres/historias_test.go`, `postgres/sprints_test.go` |
| Caso de uso | `internal/service/asignaciones.go` (`Asignaciones.Ver`, `Asignar` y `Quitar`, aviso por velocidad) | `internal/service/asignaciones_test.go` |
| Pantalla | `internal/server/planificacion.go`, `web/templates/paginas/sprint.html`, enlace en `sprints.html` | `internal/server/planificacion_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-12.1 Estimadas, con CA y no Hechas | Asignar una historia lista · Una historia de 0 SP se puede asignar · Historias que no se pueden asignar | `TestAsignarA_HistoriaLista`, `TestAsignarA_CeroStoryPointsYSprintFuturo`, `TestAsignarA_Errores`, `TestAsignar_HistoriaLista_QuedaGuardadaEnElSprint`, `TestAsignar_Rechazos_NoGuardanNada`, `TestPaginaDelSprint_MuestraElSprintYLasCandidatas`, `TestAsignar_HistoriaNoLista_MuestraElMotivo` |
| CA-12.2 Sprint abierto y un solo sprint | No se asigna a un sprint cerrado · La historia ya está en otro sprint · Planificar el sprint siguiente · Quitar una historia del sprint | `TestAsignarA_Errores` (sprint cerrado y otro sprint), `TestAsignar_YaEnOtroSprint`, `TestQuitarDelSprint_*`, `TestQuitar_*`, `TestVerSprint_CerradoYAviso` |
| CA-12.3 Aviso por velocidad | Aviso por velocidad · Sin sprints cerrados no hay aviso | `TestAsignar_AvisoPorVelocidad`, `TestAsignar_SinSprintsCerrados_NoAvisa`, `TestVerSprint_CerradoYAviso` |
