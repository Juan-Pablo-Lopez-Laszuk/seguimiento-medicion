# SPEC-28 · Gestionar el estado de un defecto

**Historia:** HU-28 (#6) · **Responsable:** Bravo, Carolina · **Sprint:** 3 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la tabla Defecto de `docs/modelo-datos.md`, la métrica M9 de
> `docs/metricas.md` y los defectos de HU-27.

## 1. Objetivo

Que un integrante pueda mover un defecto por sus estados hasta cerrarlo, indicando en qué sprint se resolvió, para
saber cuándo se resuelve cada defecto y contar los resueltos por sprint (M9, HU-34).

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto y defecto | ids (vienen en la URL) | Sí | El defecto tiene que ser del proyecto |
| Estado nuevo | `En progreso`, `Resuelto`, `Cerrado` o `Reabierto` | Sí | Una transición permitida desde el estado actual |
| Sprint de resolución | id de un sprint | Solo al pasar a Resuelto | Sprint del proyecto, igual o posterior al de detección |

## 3. Salidas esperadas

- El defecto queda en el estado nuevo y se vuelve a la lista con el mensaje **"Defecto en estado <estado>."**.
- En la lista, cada defecto muestra su sprint de resolución (si tiene) y solo los botones de las transiciones que
  valen desde su estado; el de **Resolver** viene con la lista de sprints para elegir el de resolución.

## 4. Reglas de negocio

- RN1: Las transiciones permitidas son Abierto → En progreso → Resuelto → Cerrado, Resuelto → Reabierto (CA-28.1) y
  Reabierto → En progreso, para volver a trabajarlo. Cerrado es final. Cualquier otra es un error.
- RN2: Al pasar a Resuelto se indica el sprint de resolución, que tiene que ser del proyecto y de número igual o
  mayor al del sprint de detección (CA-28.2).
- RN3: Al reabrir, se borra el sprint de resolución (CA-28.3): un defecto reabierto deja de contar como resuelto (M9).
- RN4: En las demás transiciones el sprint de resolución no cambia (Cerrado conserva el de Resuelto).

## 5. Restricciones

- Regla pura en `internal/domain/defect` (`Defecto.CambiarEstado`, `Siguientes`); caso de uso
  `service.Defectos.CambiarEstado`; repositorios con `BuscarPorID` y `Actualizar`.
- **Sin migración nueva:** `defectos.estado` y `sprint_resolucion_id` ya existen desde TEC-03.
- **Fuera de alcance:** consultar con filtros (HU-29) y contar por sprint (HU-34).

## 6. Casos límite

Sprint 1 (detección) y Sprint 2 del proyecto.

| Caso | Resultado |
|---|---|
| Abierto → En progreso → Resuelto (Sprint 2) → Cerrado | Cerrado, con Sprint 2 de resolución |
| Resolver en el Sprint 1 (el mismo de detección) | Se resuelve |
| Resolver en un sprint anterior al de detección | Error |
| Resolver sin sprint o con uno de otro proyecto | Error |
| Resuelto → Reabierto | Reabierto, sin sprint de resolución |
| Abierto → Resuelto, Abierto → Cerrado, Cerrado → Reabierto | Error: transición no permitida |
| Estado que no existe | Error |
| Defecto de otro proyecto | 404 |

## 7. Condiciones de error

| Situación | Error (Go) | Mensaje al usuario |
|---|---|---|
| Transición no permitida | `ErrTransicionInvalida` | "un defecto en estado X no puede pasar a Y" (aviso, 409) |
| Estado que no existe | `ErrEstadoInvalido` | "el estado tiene que ser En progreso, Resuelto, Cerrado o Reabierto" (400) |
| Resuelto sin sprint o de otro proyecto | `ErrSprintResolucionInvalido` | "el sprint de resolución tiene que ser un sprint del proyecto" (aviso, 409) |
| Sprint de resolución anterior al de detección | `ErrSprintResolucionAnterior` | "el sprint de resolución no puede ser anterior al de detección" (aviso, 409) |
| Defecto o proyecto inexistente | `ErrDefectoNoEncontrado` / `project.ErrNoEncontrado` | Página 404 |

## 8. Criterios de aceptación

- CA-28.1: Abierto → En progreso → Resuelto → Cerrado, y Resuelto → Reabierto.
- CA-28.2: Al resolverlo se indica el sprint de resolución (igual o posterior al de detección).
- CA-28.3: Si se reabre, se borra el sprint de resolución.

## Trazabilidad

Issue: #6 · Métrica: M9 · Feature: `features/hu-28-estado-defecto.feature` · Pasos BDD: `features/hu28_steps_test.go`
