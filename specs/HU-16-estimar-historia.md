# SPEC-16 · Estimar historia con story points

**Historia:** HU-16 (#24) · **Responsable:** Piastrellini, Mariano · **Sprint:** 2 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la columna `historias.story_points` de `docs/modelo-datos.md`, la escala
> de story points de HU-05 (RN4) y la decisión de Juan Pablo del 15/10 de dividir esta historia: el historial de
> estimaciones (CA-16.3) pasó a **HU-16b (#106)**.

## 1. Objetivo

Que el equipo pueda estimar una historia del backlog con story points, o cambiar una estimación que ya tenía, siempre
con un valor de la escala acordada. HU-05 ya permite cargar los story points al crear la historia; esta historia suma
cambiarlos después y deja la regla en un solo lugar (`Historia.Estimar`) para que la usen también las pantallas de
Planning Poker (HU-17 a HU-19) cuando el equipo acuerde un valor.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Historia | número de la historia (HU-N, viene en la URL) | Sí | Tiene que existir en ese proyecto |
| Story points | número | Sí | Uno de la escala 0, 1, 2, 3, 5, 8, 13, 21 |

## 3. Salidas esperadas

- **Si el valor vale:** la historia queda con los story points nuevos, guardados. Se vuelve a su página con el mensaje
  **"Estimación guardada."** y el valor nuevo visible.
- **Si no vale:** la historia **no cambia** y el error aparece debajo del campo (RN7).
- **Página de la historia** (`/proyectos/{id}/backlog/{numero}`): junto a los story points hay un desplegable con la
  escala, con el valor actual elegido, y el botón **"Guardar estimación"**. Si la historia está Hecha no se ofrece, y la
  página lo explica.

## 4. Reglas de negocio

- RN1: Los story points tienen que ser **uno de la escala acordada**: 0, 1, 2, 3, 5, 8, 13 o 21 (CA-16.1). Es la misma
  escala (`backlog.EscalaStoryPoints`) que valida HU-05 al crear la historia y la base (`historias_story_points_ck`).
- RN2: Un valor **fuera de la escala**, o que no es un número entero, es un error del campo `story_points` (CA-16.2), con el
  mismo mensaje que al crear la historia.
- RN3: El valor es **obligatorio**: vacío es un error. Esta historia estima o re-estima; no deja una historia "sin
  estimar" otra vez ("sin estimar" es solo cómo nace una historia sin story points).
- RN4: **0 cuenta como estimada** y "sin estimar" no (igual que HU-05 RN4 y HU-12 RN1).
- RN5: **Re-estimar** reemplaza el valor anterior por el nuevo. Pedir el mismo valor que ya tiene no es un error: queda
  igual. El valor anterior **no se conserva**: guardarlo es HU-16b (CA-16.3).
- RN6: Se puede estimar una historia mientras no esté Hecha: en **Pendiente, En sprint o En progreso**. Una **historia Hecha
  no se re-estima**, porque los story points completados de su sprint (HU-30 y HU-32) cambiarían después de haberse
  terminado.
- RN7: Si el cambio no vale, la historia **queda como estaba**: no se guarda nada.
- RN8: Estimar **no toca** el estado, los criterios, las tareas ni el sprint de la historia.

## 5. Restricciones

- La regla es una **función pura** en `internal/domain/backlog` (`Historia.Estimar`): recibe el texto del valor y devuelve
  la historia con los story points nuevos o el error. No usa base de datos ni HTTP, así se prueba con TDD. Reusa la
  validación de la escala de HU-05.
- El caso de uso en `internal/service` (`Estimaciones.Estimar`) busca el proyecto y la historia, llama a la regla y guarda.
  **Cambios en los repositorios:** el de historias suma `ActualizarStoryPoints` (memoria y PostgreSQL), que es el
  **único lugar donde se escribe el valor**; HU-16b lo va a extender para anotar también el historial.
- **Sin migración nueva:** la columna `historias.story_points` ya existe y su `CHECK` ya acepta la escala (decisión de
  Juan Pablo: HU-16 sin historial no trae migración, como asumió la Planning).
- Se estima con un formulario `POST` y una redirección 303, como en el resto de la aplicación. La pantalla usa los
  componentes de TEC-07 (`exito`, `aviso`, `errorCampo`).
- **Fuera de alcance**, porque tienen su propia historia: el historial de estimaciones (HU-16b), quitar una estimación,
  y las sesiones de Planning Poker (HU-17 a HU-19), que van a usar `Historia.Estimar` para guardar el valor acordado.

## 6. Casos límite

Historia de ejemplo: HU-1 "Iniciar sesión con Google".

| Caso | Resultado |
|---|---|
| Historia sin estimar, se estima en 5 | Queda con 5 story points |
| Historia con 5, se estima en 8 | Queda con 8 (reemplaza al anterior) |
| Historia con 5, se pide 5 | Queda con 5, sin error |
| Los ocho valores de la escala (0, 1, 2, 3, 5, 8, 13, 21) | Valen |
| 0 | Vale, y la historia ya no está "sin estimar" |
| 4, 22, -1, "3,5", "abc" | Error de story points |
| Valor vacío | Error de story points |
| Un valor inválido sobre una historia ya estimada | Error, y la estimación anterior se conserva |
| Historia Pendiente, En sprint o En progreso | Se puede estimar |
| Historia Hecha | Error: no se puede cambiar su estimación (409) |
| Proyecto o historia que no existe | Página "no encontrado" (404) |

## 7. Condiciones de error

| Situación | Campo | Error (Go) | Mensaje al usuario |
|---|---|---|---|
| Valor fuera de la escala, vacío o que no es un número | story_points | `ErrStoryPointsInvalidos` (de HU-05) | "los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21" |
| La historia está Hecha | — | `ErrEstimacionDeHistoriaHecha` | "la historia está Hecha: no se puede cambiar su estimación" (aviso en la página, 409) |
| El proyecto no existe | — | `project.ErrNoEncontrado` | Página 404 "No encontramos ese proyecto" |
| La historia no existe en el proyecto | — | `backlog.ErrHistoriaNoEncontrada` | Página 404 "No encontramos esa historia" |

En la pantalla, el error del valor aparece debajo del campo.

## 8. Criterios de aceptación

- CA-16.1: Solo valores de la escala acordada (Fibonacci: 0, 1, 2, 3, 5, 8, 13, 21).
- CA-16.2: Un valor fuera de la escala devuelve error.
- ~~CA-16.3~~ Re-estimar conserva el valor anterior en el historial: pasó a **HU-16b (#106)**.

## Trazabilidad

Issue: #24 · Feature: `features/hu-16-estimar-historia.feature` · Pasos BDD: `features/hu16_steps_test.go`

Se completa al terminar la implementación.
