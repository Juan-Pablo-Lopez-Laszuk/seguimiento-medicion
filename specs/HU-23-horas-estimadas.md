# SPEC-23 · Registrar horas estimadas

**Historia:** HU-23 (#1) · **Responsable:** Bravo, Carolina · **Sprint:** 1 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`, la métrica M4 de `docs/metricas.md` y las columnas `historias.horas_estimadas` y `tareas.horas_estimadas` de `docs/modelo-datos.md`.

## 1. Objetivo

Que el equipo pueda cargar cuántas horas espera dedicarle a una historia, para después compararlas con las horas reales (HU-24) y calcular la desviación (HU-33).

**Alcance en el Sprint 1 (Planning #85):** se hace **la regla** con TDD. Guardar las horas en la historia espera a la Historia de HU-05 (Mariano); la pantalla se agrega cuando exista.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Horas que escribe el usuario | texto | Sí | Número mayor a 0, con coma o punto decimal ("1,5" o "1.5") |
| Horas cargadas en la historia | número | No (0 = sin estimar) | ≥ 0 |
| Tareas de la historia, con sus horas | lista | No (puede estar vacía) | Horas de cada tarea ≥ 0 (0 = sin estimar) |

## 3. Salidas esperadas

- Las horas como número (`float64`), listas para guardar.
- Las horas estimadas de la historia: la suma de sus tareas, o las de la historia si no tiene tareas.

## 4. Reglas de negocio

- RN1: Las horas estimadas son mayores a 0 y aceptan decimales (CA-23.1).
- RN2: Se acepta coma o punto decimal y espacios alrededor ("1,5", "1.5", " 0,25 ").
- RN3: Si la historia tiene tareas, sus horas estimadas son la suma de las horas de sus tareas; las horas cargadas en la historia se ignoran (CA-23.2). Una sola fuente de verdad, como dice `docs/modelo-datos.md`.
- RN4: Si la historia no tiene tareas, se usan las horas cargadas en la historia.
- RN5: Una historia o tarea todavía sin estimar vale 0 en la suma.

## 5. Restricciones

- Reglas puras en `internal/domain/effort` (sin base de datos ni HTTP), así se prueban con TDD sin depender de nadie.
- Internamente `float64`.
- La base tiene el mismo control: `CHECK horas_estimadas > 0` cuando no es nulo (`docs/modelo-datos.md`).

## 6. Casos límite

| Caso | Resultado |
|---|---|
| "0,25" | 0,25 h (decimales chicos se aceptan) |
| "1.5" con punto | 1,5 h |
| Historia sin tareas y sin estimar | 0 |
| Historia con 10 h cargadas y tareas de 2 h y 3 h | 5 h (manda la suma de tareas) |

## 7. Condiciones de error

| Situación | Error | Mensaje al usuario |
|---|---|---|
| Horas en 0 o negativas | `ErrHorasInvalidas` | "las horas estimadas deben ser mayores a 0" |
| Texto vacío o que no es un número ("abc", "1,5,2") | `ErrHorasInvalidas` | "las horas estimadas deben ser mayores a 0" |
| Una tarea o la historia con horas negativas | `ErrHorasInvalidas` | "las horas estimadas deben ser mayores a 0" |

## 8. Criterios de aceptación

- CA-23.1: Las horas son mayores a 0 y aceptan decimales (ej. 1,5).
- CA-23.2: Si la historia tiene tareas, se suman las horas de las tareas.

## Trazabilidad

Issue: #1 · Métrica: M4 en `docs/metricas.md` · Feature: `features/hu-23-horas-estimadas.feature` · Pasos BDD: `features/hu23_steps_test.go` · Tests: `internal/domain/effort/horas_test.go` · Código: `internal/domain/effort/horas.go` (`ValidarHoras`, `ParsearHoras`, `HorasEstimadasDeHistoria`)

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-23.1 | Cargar horas en una historia sin tareas · Horas con decimales · Horas inválidas | `TestValidarHoras_MayoresACeroConDecimales`, `TestParsearHoras_AceptaComaOPuntoDecimal`, `TestParsearHoras_RechazaTextosInvalidos` |
| CA-23.2 | Historia con tareas | `TestHorasEstimadasDeHistoria_SumaLasTareas`, `TestHorasEstimadasDeHistoria_HorasNegativasSonError` |
