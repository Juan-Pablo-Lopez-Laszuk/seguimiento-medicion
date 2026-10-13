# SPEC-13 · Registrar historias completadas

**Historia:** HU-13 (#45) · **Responsable:** López, Juan Pablo · **Sprint:** 2 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la página del sprint de `specs/HU-12-asignar-historias.md`, el
> sprint activo de `specs/HU-04-estado-proyecto.md` (RN2) y la regla de estados de HU-08 (Mariano). Lo que viene de HU-08
> está marcado y hay que acordarlo con él antes de programar (decisión 6 de la Planning).

## 1. Objetivo

Que el Product Builder marque como completada una historia del sprint activo, así queda registrado el avance y la fecha
en que se terminó (la usan las métricas y el burndown).

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Sprint | número del sprint (viene en la URL) | Sí | Tiene que ser el sprint activo (RN1) |
| Historia | id de la historia (viene en el formulario) | Sí | Tiene que estar en ese sprint, En progreso y con todos sus criterios cumplidos (RN2 y RN3) |
| Hoy | fecha | — | No la carga el usuario: es la fecha de Argentina (HU-04 RN6) |

## 3. Salidas esperadas

- En la página del sprint (HU-12), cada historia **En progreso** del sprint activo tiene el botón **Completar**, y cada
  historia **En sprint** tiene el botón **Empezar** (la transición En sprint → En progreso de HU-08).
- Al completar, la historia pasa a **Hecho**, se guarda la **fecha de completado** (`completada_en`, hoy en Argentina)
  y se vuelve a la página del sprint con el aviso "HU-N completada".
- La página muestra el avance del sprint: **SP completados de SP comprometidos** (por ejemplo "8 de 13 SP"), con
  `metrics.CalcularStoryPoints` de HU-30.
- Si a la historia le faltan criterios por cumplir, se muestra cuántos (por ejemplo "faltan 1 de 3 criterios") en lugar
  del botón.

## 4. Reglas de negocio

- RN1: **Solo historias del sprint activo** (CA-13.1). El sprint activo es el de HU-04 RN2. En un sprint futuro o
  cerrado no se completa nada.
- RN2: **Hay que pasar por En progreso** (CA-08.1: Pendiente → En sprint → En progreso → Hecho). Una historia En sprint
  primero se empieza y después se completa. **A confirmar con Mariano (decisión 6).**
- RN3: **Todos los criterios cumplidos** (CA-13.3 y CA-08.2). La regla es la de HU-08; HU-13 la usa, no la repite.
- RN4: **Fecha de completado** (CA-13.2): se guarda hoy, en la hora de Argentina. No la elige el usuario.
- RN5: **No se deshace.** Una historia Hecha no vuelve atrás desde esta pantalla: CA-08.1 solo prevé la vuelta a
  Pendiente al cerrar el sprint (HU-14). **A confirmar.**
- RN6: Completar una historia no cambia los SP comprometidos del sprint, solo los completados.

## 5. Restricciones

- La transición y la regla de criterios son de HU-08 (`internal/domain/backlog`). HU-13 agrega la condición del sprint
  activo y la fecha.
- `backlog.Historia` necesita el campo de la fecha de completado, y el `Actualizar` del repositorio de historias (el
  mismo que se acuerda para HU-08 y HU-12) tiene que guardarla. `historias.completada_en` ya está en la migración 0001.
- HU-13 se programa después de HU-12 y HU-08.

## 6. Casos límite

Sprint 2 activo, del 12/10 al 18/10. Hoy es el 15/10.

| Caso | Resultado |
|---|---|
| Completar HU-1 (En progreso, 2 de 2 criterios cumplidos) | Hecho, completada el 15/10/2026 |
| Completar HU-2 (En progreso, 1 de 2 criterios cumplidos) | Error: le falta 1 criterio |
| Completar HU-3 (En sprint, todavía no empezó) | Error: primero hay que empezarla (RN2) |
| Empezar HU-3 | Pasa a En progreso |
| Completar una historia del Sprint 3 (todavía no empezó) | Error: no es el sprint activo |
| Completar una historia ya Hecha | Error: ya está Hecha |
| A las 22:00 de Argentina del 18/10 (en UTC ya es 19/10) | Se completa con fecha 18/10/2026 (RN4) |
| Sprint 2 vencido sin cerrar (hoy 20/10) | Sigue activo (HU-04 RN2): se puede completar, con fecha 20/10/2026 |
| Avance: HU-1 (5 SP) Hecha de 13 SP comprometidos | "5 de 13 SP" |

## 7. Condiciones de error

| Situación | Error (Go) | Mensaje al usuario |
|---|---|---|
| El sprint no es el activo | `ErrNoEsElSprintActivo` | "solo se completan historias del sprint activo" |
| La historia no está En progreso | la de HU-08 | "HU-3 todavía no empezó: primero hay que pasarla a En progreso" |
| Faltan criterios por cumplir | la de HU-08 | "HU-2 no se puede completar: falta cumplir 1 de 2 criterios" |
| Ya está Hecha | la de HU-08 | "HU-1 ya está Hecha" |
| La historia no está en ese sprint, o no existe | — | Página 404 |

Los errores se muestran como aviso en la página del sprint (409). Los textos exactos de los errores de HU-08 los define
Mariano; los de esta tabla son la propuesta.

## 8. Criterios de aceptación

- CA-13.1: Sólo historias del sprint activo.
- CA-13.2: Se guarda la fecha de completado.
- CA-13.3: Exige los CA cumplidos (regla de HU-08).

## Trazabilidad

Issue: #45 · Feature: `features/hu-13-completar-historias.feature` · Pasos BDD: `features/hu13_steps_test.go`

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-13.1 Solo el sprint activo | Completar una historia del sprint activo · No se completa fuera del sprint activo · Sprint vencido sin cerrar | se completa con el código |
| CA-13.2 Fecha de completado | Completar una historia del sprint activo · La fecha es la de Argentina | se completa con el código |
| CA-13.3 Criterios cumplidos | Faltan criterios · Hay que empezarla antes | se completa con el código |
