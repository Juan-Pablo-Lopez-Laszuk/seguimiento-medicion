# SPEC-24 · Registrar esfuerzo realizado

**Historia:** HU-24 (#2) · **Responsable:** Bravo, Carolina · **Sprint:** 2 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`, la tabla RegistroEsfuerzo de
> `docs/modelo-datos.md`, la métrica M5 de `docs/metricas.md` y las tareas de HU-25.

## 1. Objetivo

Que cada integrante pueda cargar las horas que trabajó en una historia (o en una de sus tareas), con la fecha y la
actividad, para medir el esfuerzo real (M5) y después compararlo con lo estimado (HU-33).

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto e historia | id del proyecto y número HU-N (vienen en la URL) | Sí | Tienen que existir |
| Integrante | id de un integrante | Sí | Integrante **activo** del proyecto (CA-24.3) |
| Tarea | id de una tarea | No | Si viene, tiene que ser de esa historia (HU-25) |
| Fecha | `AAAA-MM-DD` | Sí | No futura y dentro de las fechas del proyecto (CA-24.2) |
| Actividad | texto | Sí | Hasta 200 caracteres, sin espacios sobrantes |
| Horas | texto | Sí | Mayor a 0 y hasta 24, con coma o punto (CA-24.1) |

## 3. Salidas esperadas

- El registro queda guardado (integrante, historia, tarea si se eligió, fecha, actividad y horas) y se vuelve a la
  página de esfuerzo de la historia con el mensaje **"Horas registradas."**.
- **Página de esfuerzo** (`/proyectos/{id}/backlog/{numero}/esfuerzo`): los registros de la historia (fecha más
  reciente primero) con integrante, tarea, actividad y horas; el **total de horas reales** de la historia (M5); y el
  formulario. Se llega desde la página de la historia y desde la de tareas.

## 4. Reglas de negocio

- RN1: Las horas de un registro son mayores a 0 y hasta 24, con coma o punto decimal (CA-24.1).
- RN2: Un integrante no suma más de 24 h en un mismo día, contando todos sus registros de ese día en cualquier
  historia (CA-24.1).
- RN3: La fecha no puede ser posterior a hoy (hora de Argentina) ni estar fuera de las fechas del proyecto (CA-24.2).
- RN4: El integrante tiene que ser un integrante activo del proyecto (CA-24.3).
- RN5: El registro es de la historia de la URL; la tarea es opcional y, si se elige, tiene que ser de esa historia.
- RN6: La actividad es obligatoria, hasta 200 caracteres, sin los espacios del inicio y del final.
- RN7: Las horas reales de una historia son la suma de sus registros (M5).
- RN8: Todos los errores de los campos se informan juntos.

## 5. Restricciones

- Reglas puras en `internal/domain/effort` (`NuevoRegistro`, `HorasReales`); el caso de uso en `internal/service`;
  repositorios en `internal/store/memory` y `internal/store/postgres` (`Guardar`, `ListarPorHistoria`,
  `ListarPorIntegrante`).
- **Sin migración nueva:** la tabla `registros_esfuerzo` ya existe desde TEC-03, con `CHECK horas > 0 AND horas <= 24`.
  El registro guarda siempre `historia_id`, y `tarea_id` solo si se eligió una tarea.
- **Fuera de alcance:** consultar y corregir registros (HU-26) y la desviación estimado/real (HU-33).

## 6. Casos límite

Proyecto del 05/10/2026 al 01/11/2026, hoy 18/10/2026. Carolina activa, Ana dada de baja.

| Caso | Resultado |
|---|---|
| Carolina, 18/10, "Pantalla de login", 3 h | Se registra; la historia suma 3 h reales |
| 24 h en un registro / 24,5 h / 0 h | Se registra / error / error |
| Carolina ya tiene 20 h el 17/10 y carga 4 / 5 | Se registra (24) / error de total diario |
| Fecha 19/10 (mañana) | Error: fecha futura |
| Fecha 04/10 (antes del proyecto) | Error: fuera del proyecto |
| Ana (dada de baja) o un integrante de otro proyecto | Error: integrante del proyecto |
| Tarea de otra historia | Error: la tarea no es de la historia |
| Actividad vacía o de 201 caracteres | Error |

## 7. Condiciones de error

| Situación | Campo | Error (Go) | Mensaje al usuario |
|---|---|---|---|
| Horas en 0, más de 24 o no número | horas | `ErrHorasTrabajadasInvalidas` | "las horas tienen que ser mayores a 0 y hasta 24" |
| Supera 24 h en el día | horas | `ErrMasDe24HorasEnElDia` | "el integrante no puede superar 24 h en un día" |
| Fecha vacía o inválida | fecha | `ErrFechaInvalida` | "la fecha es obligatoria" |
| Fecha futura | fecha | `ErrFechaFutura` | "la fecha no puede ser futura" |
| Fecha fuera del proyecto | fecha | `ErrFechaFueraDelProyecto` | "la fecha tiene que estar dentro de las fechas del proyecto" |
| Integrante vacío, de baja o de otro proyecto | integrante | `ErrIntegranteInvalido` | "el integrante tiene que ser un integrante activo del proyecto" |
| Tarea de otra historia | tarea | `ErrTareaDeOtraHistoria` | "la tarea tiene que ser de esta historia" |
| Actividad vacía / larga | actividad | `ErrActividadObligatoria` / `ErrActividadLarga` | "la actividad es obligatoria" / "la actividad no puede superar los 200 caracteres" |
| Proyecto o historia inexistente | — | `project.ErrNoEncontrado` / `backlog.ErrHistoriaNoEncontrada` | Página 404 |

## 8. Criterios de aceptación

- CA-24.1: Entre 0 y 24 h por registro, y no más de 24 h por día.
- CA-24.2: La fecha no puede ser futura ni estar fuera del proyecto.
- CA-24.3: El integrante tiene que ser del proyecto.

## Trazabilidad

Issue: #2 · Métrica: M5 · Feature: `features/hu-24-registro-esfuerzo.feature` · Pasos BDD: `features/hu24_steps_test.go`
