# SPEC-16b · Historial de estimaciones

**Historia:** HU-16b (#106) · **Responsable:** Piastrellini, Mariano · **Sprint:** 3 · **Estado:** Revisada

> Especificación escrita antes del código (SDD). Sale de dividir HU-16 (decisión de Juan Pablo del 15/10, issue #106):
> HU-16 (`specs/HU-16-estimar-historia.md`) quedó con CA-16.1 y CA-16.2, y esta historia suma CA-16.3. Usa
> `Historia.Estimar` y `ActualizarStoryPoints` de HU-16, y trae la **migración 0003**.

## 1. Objetivo

Que al re-estimar una historia se conserve el valor anterior, para que el equipo vea cómo cambió su estimación. Hoy
`ActualizarStoryPoints` reemplaza el valor y el anterior se pierde.

## 2. Entradas

Esta historia no suma ningún dato que cargar: el historial se genera solo, cuando se estima una historia desde la
pantalla de HU-16 (o, más adelante, desde Planning Poker).

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Historia | número de la historia (HU-N, viene en la URL) | Sí | Tiene que existir en ese proyecto |
| Story points nuevos | número | Sí | Los de HU-16 (escala 0, 1, 2, 3, 5, 8, 13, 21) |

## 3. Salidas esperadas

- **Al re-estimar con un valor distinto:** la historia queda con el valor nuevo, igual que en HU-16, y se agrega al
  historial una entrada con el **valor anterior**, el **valor nuevo** y **cuándo** se cambió.
- **Página de la historia** (`/proyectos/{id}/backlog/{numero}`): debajo de la estimación hay una sección
  **"Historial de estimaciones"** con una línea por cambio, de la más reciente a la más antigua:
  `20/10/2026 18:30 · 5 → 8` (fecha y hora de Argentina). Si todavía no se re-estimó, dice
  **"Todavía no se re-estimó esta historia."**
- El historial **también se ve en una historia Hecha** (solo se consulta; ahí no se re-estima, RN6 de HU-16).

## 4. Reglas de negocio

- RN1: **Re-estimar anota el cambio.** Si la historia ya tenía una estimación y el valor nuevo es distinto, se agrega al
  historial una entrada con el valor anterior, el valor nuevo y la fecha y hora del cambio (CA-16.3).
- RN2: **La primera estimación no se anota**: pasar de "sin estimar" a un valor no tiene valor anterior que conservar. Una
  historia que nace con story points (HU-05) tampoco tiene historial hasta que se re-estima.
- RN3: **Pedir el mismo valor no se anota**: no cambió nada (RN5 de HU-16).
- RN4: **0 es una estimación** (RN4 de HU-16): pasar de 0 a 5, o de 5 a 0, se anota. "Sin estimar" no es un valor.
- RN5: **Un cambio que no vale no anota nada.** Un valor inválido, o una historia Hecha, dejan la estimación y el
  historial como estaban (RN6 y RN7 de HU-16).
- RN6: **El historial solo crece**: la aplicación agrega entradas y no las edita ni las borra (decisión de Juan Pablo).
  La única excepción es RN9.
- RN7: **El historial es de cada historia**: re-estimar HU-1 no toca el de HU-2.
- RN8: **Todo o nada**: el cambio de la estimación y su entrada del historial se guardan juntos. Si falla uno, no queda
  guardado ninguno.
- RN9: **Al eliminar una historia (HU-06) se elimina su historial**, igual que sus criterios y sus tareas: un historial
  sin historia no tiene sentido y la base no lo borra sola (TEC-03).
- RN10: **No se guarda quién estimó**, porque todavía no hay login. Cuando llegue TEC-05 se le suma la columna.
- RN11: **Las historias estimadas antes de la migración** empiezan con el historial vacío; su primer cambio ya conserva el
  valor anterior, sin necesidad de rellenar nada.

## 5. Restricciones

- **La regla de qué se anota es una función pura** en `internal/domain/backlog`, que recibe la estimación anterior (puede
  ser "sin estimar"), la nueva y la hora, y devuelve la entrada a anotar o que no corresponde (RN1 a RN4). No usa base
  de datos ni HTTP, así se prueba con TDD. Los dos repositorios la usan, así que la regla está en un solo lugar.
- **Migración 0003** (`migrations/0003_estimaciones_historia.sql`, en su propio PR, sin el `CHECK` de fechas de `sprints`
  que va en la 0004): tabla `estimaciones_historia` con `id`, `historia_id` (FK a `historias`), `valor_anterior`,
  `valor_nuevo` y `cambiada_en` (`timestamptz`), con los mismos `CHECK` de escala que `historias` y `valor_anterior <>
  valor_nuevo`. Sin `ON DELETE CASCADE`, como el resto del modelo. **RLS habilitado y sin políticas**, como las demás
  tablas (`0001` y `0002`): Supabase publica por API cada tabla de `public`. Se actualiza `docs/modelo-datos.md`.
- **`ActualizarStoryPoints` sigue siendo el único lugar donde se escribe la estimación** (HU-16): pasa a recibir la hora
  del cambio. En PostgreSQL hace todo en **una transacción**: lee los story points actuales con `FOR UPDATE`, los
  actualiza y agrega la entrada si corresponde, así dos re-estimaciones a la vez no pierden un valor. En memoria hace lo
  mismo con su candado.
- **Lectura:** los repositorios suman `ListarEstimaciones(ctx, historiaID)`, de la más reciente a la más antigua.
- **Caso de uso:** `Estimaciones` recibe el reloj (`ahora func() time.Time`, como `Asignaciones`) y suma `Historial` (con
  el proyecto y el número de la historia). La página de la historia lo muestra.
- **HU-06:** `Eliminar` borra también el historial dentro de su transacción (PostgreSQL) y el repositorio en memoria
  suma el historial a sus dependientes (RN9).
- **Sin cambios** en HU-05 (crear), en HU-06 (editar) ni en el formulario de estimar de HU-16. La pantalla usa los
  componentes de TEC-07.
- **Producción:** la migración 0003 la aplica Juan Pablo el mismo día del merge (acción de la Retro del Sprint 1); sin
  ella, el deploy de Vercel queda con el esquema viejo.
- **Fuera de alcance**, porque tienen su propia historia o no se pidió: quién estimó (TEC-05), un gráfico de la evolución,
  corregir o borrar entradas, y el historial de otros campos. Las sesiones de Planning Poker (HU-17 a HU-22) van a
  guardar el valor acordado con `Historia.Estimar`, y por eso van a dejar su historial sin código nuevo.

## 6. Casos límite

Historia de ejemplo: HU-1 "Iniciar sesión con Google".

| Caso | Resultado |
|---|---|
| Sin estimar, se estima en 5 | Queda con 5, historial vacío |
| Estimada en 5, se re-estima en 8 | Queda con 8, historial: `5 → 8` |
| Estimada en 5, se re-estima en 8 y después en 3 | Historial, el más reciente primero: `8 → 3`, `5 → 8` |
| Estimada en 5, se pide 5 | Queda con 5, historial vacío, sin error |
| Estimada en 0, se re-estima en 5 | Historial: `0 → 5` |
| Estimada en 5, se re-estima en 0 | Historial: `5 → 0` |
| Estimada en 5, se pide un valor inválido (4, vacío, "abc") | Error de story points, queda con 5 y el historial no cambia |
| Historia Hecha | No se re-estima (409): el historial no cambia, pero se ve |
| Se re-estima HU-1 | El historial de HU-2 no cambia |
| Se elimina una historia con historial | Se elimina la historia, con su historial |
| Historia que nunca se re-estimó | "Todavía no se re-estimó esta historia." |
| Proyecto o historia que no existe | Página "no encontrado" (404) |

## 7. Condiciones de error

No hay errores nuevos: valen los de HU-16 (valor inválido, historia Hecha, proyecto o historia inexistente).

| Situación | Error (Go) | Respuesta |
|---|---|---|
| Valor fuera de la escala, vacío o que no es un número | `ErrStoryPointsInvalidos` (de HU-05) | Error debajo del campo; no se anota nada |
| La historia está Hecha | `ErrEstimacionDeHistoriaHecha` (de HU-16) | Aviso en la página (409); no se anota nada |
| Falla al guardar el historial | error del repositorio | 500; la estimación no cambia (RN8) |

## 8. Criterios de aceptación

- CA-16.3: Re-estimar conserva el valor anterior en el historial.

## Trazabilidad

Issue: #106 · Feature: `features/hu-16b-historial-estimaciones.feature` · Pasos BDD: `features/hu16b_steps_test.go`

Las columnas de código y tests se completan al implementar.

| Capa | Código | Tests |
|---|---|---|
| Dominio | `internal/domain/backlog/` (regla de qué se anota) | |
| Migración | `migrations/0003_estimaciones_historia.sql` | |
| Repositorio en memoria | `ActualizarStoryPoints` y `ListarEstimaciones` en `internal/store/memory/historias.go` | |
| Repositorio PostgreSQL | `ActualizarStoryPoints`, `ListarEstimaciones` y `Eliminar` en `internal/store/postgres/historias.go` | |
| Caso de uso | `internal/service/estimaciones.go` (`Estimaciones.Estimar` y `Estimaciones.Historial`) | |
| Pantalla | `web/templates/paginas/historia.html` | |
| Aplicación | `app/app.go` | |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-16.3 Re-estimar conserva el valor anterior | Re-estimar conserva el valor anterior · Varias re-estimaciones, la más reciente primero · La primera estimación no tiene valor anterior · Pedir el mismo valor no anota nada · Cero cuenta como estimación · Un valor inválido no anota nada · Una historia Hecha no se re-estima · El historial es de cada historia | |
| RN9 Eliminar una historia elimina su historial | Eliminar una historia con historial | |
