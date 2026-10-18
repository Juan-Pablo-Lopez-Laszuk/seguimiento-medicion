# SPEC-25 · Descomponer historia en tareas

**Historia:** HU-25 (#3) · **Responsable:** Bravo, Carolina · **Sprint:** 2 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`, la tabla Tarea de
> `docs/modelo-datos.md` y la regla de horas de HU-23 (`specs/HU-23-horas-estimadas.md`).

## 1. Objetivo

Que el Product Builder pueda dividir una historia del backlog en tareas, cada una con su responsable y sus horas
estimadas, para organizar el trabajo. Las tareas son la base de HU-24 (registrar horas trabajadas en una tarea) y,
como dice CA-23.2, cuando una historia tiene tareas sus horas estimadas pasan a ser la suma de las de sus tareas.

## 2. Entradas

**Agregar una tarea**

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Historia | número de la historia (HU-N, viene en la URL) | Sí | Tiene que existir en ese proyecto |
| Título | texto | Sí | Hasta 120 caracteres (después de quitar espacios al inicio y al final) |
| Responsable | id de un integrante | Sí | Integrante **activo** del mismo proyecto |
| Horas estimadas | texto | Sí | Número mayor a 0, con coma o punto decimal (regla de HU-23) |

**Cambiar las horas estimadas de una tarea**

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto, historia y tarea | ids (vienen en la URL) | Sí | La tarea tiene que ser de esa historia |
| Horas estimadas | texto | Sí | Número mayor a 0, con coma o punto decimal |

**Marcar una tarea**

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto, historia y tarea | ids (vienen en la URL) | Sí | La tarea tiene que ser de esa historia |
| Completada | `1` (completada) o `0` (pendiente) | Sí | Cualquier otro valor es un error |

La historia de la tarea **no se elige**: es la de la URL (RN4). Toda tarea nace pendiente (RN5).

## 3. Salidas esperadas

- **Al agregar:** la tarea queda guardada con id, historia, título, responsable, horas y **pendiente**. Se vuelve a la
  página de tareas con el mensaje **"Tarea agregada."** y la tarea nueva al final de la lista.
- **Al cambiar las horas:** se vuelve a la página con el mensaje **"Horas estimadas actualizadas."**.
- **Al marcar:** se vuelve a la página con **"Tarea marcada como completada."** o **"Tarea marcada como pendiente."**.
- **Página de tareas** (`/proyectos/{id}/backlog/{numero}/tareas`): la historia (HU-N y título), sus tareas en orden
  de carga con responsable, horas y estado, y las **horas estimadas de la historia** (la suma de sus tareas, CA-23.2).
  Cada tarea pendiente tiene un formulario para cambiar sus horas y un botón para completarla; una completada, solo el
  botón para volver a pendiente. Debajo, el formulario para agregar una tarea, con los integrantes activos del
  proyecto para elegir el responsable. Sin tareas dice "Sin tareas todavía".
- La página de la historia (HU-07) tiene un enlace a sus tareas.

## 4. Reglas de negocio

- RN1: El título se guarda **sin los espacios del inicio y del final**; si queda vacío, es obligatorio. El largo se
  cuenta en caracteres (como el de una historia, hasta 120).
- RN2: El responsable tiene que ser un **integrante activo del proyecto** de la historia. Uno dado de baja o de otro
  proyecto no puede ser responsable de una tarea nueva (CA-25.1).
- RN3: Las horas estimadas son mayores a 0 y aceptan coma o punto decimal: es la regla de HU-23 (CA-23.1), la misma
  función `effort.ParsearHoras`.
- RN4: Una tarea pertenece a **una sola historia** (CA-25.2): la historia se fija al crearla y no se cambia. Las tareas
  se muestran, se suman y se modifican solo dentro de su historia; una tarea con el id de otra historia es un "no
  encontrado".
- RN5: Toda tarea nace **pendiente** (no completada).
- RN6: Una tarea **completada no cambia su estimación** (CA-25.3): querer cambiarle las horas es un error.
- RN7: Una tarea se marca completada o pendiente **de forma explícita** (el pedido dice a cuál de los dos estados
  pasa), igual que los criterios de HU-07: un doble clic no la deja al revés y marcarla dos veces igual no es un error.
  Volver a pendiente sirve para corregir un error; recién ahí se le pueden cambiar las horas.
- RN8: Las **horas estimadas de la historia** son la suma de las horas de sus tareas (CA-23.2,
  `effort.HorasEstimadasDeHistoria`). Las tareas completadas también suman: la estimación no cambia por terminarla.
- RN9: Las tareas se muestran **en el orden en que se cargaron**.

## 5. Restricciones

- Las reglas de una tarea (RN1 a RN3, RN5 a RN7) son **funciones puras** en `internal/domain/effort` (épica E5, junto
  a las horas de HU-23): reciben los datos y los integrantes del proyecto, y devuelven la tarea o el error. La
  `effort.Tarea` de HU-23, que solo tenía las horas, suma los demás campos.
- El caso de uso en `internal/service` busca el proyecto, la historia (con `BuscarPorNumero` de HU-07) y los
  integrantes, llama al dominio y guarda. **Repositorio nuevo de tareas** (`ListarPorHistoria`, `Guardar`,
  `BuscarPorID` y `Actualizar`), en `internal/store/memory` para los tests y en `internal/store/postgres` para la
  aplicación.
- **Sin migración nueva:** la tabla `tareas` ya existe desde TEC-03 (`id`, `historia_id`, `titulo`, `responsable_id`,
  `horas_estimadas`, `completada`) con el `CHECK horas_estimadas > 0`.
- Agregar, cambiar horas y marcar se piden con un formulario `POST` y se vuelve a la página con una redirección 303,
  como en el resto de la aplicación. Las pantallas usan los componentes de TEC-07 (`exito`, `aviso`, `error`,
  `errorCampo`).
- Las tareas tienen una página propia y la página de la historia solo suma un enlace, para no pisar el trabajo de
  HU-08 sobre esa página.
- Las tareas no tienen `ON DELETE CASCADE` (decisión de TEC-03): qué pasa con ellas al eliminar una historia lo definió
  HU-06: se eliminan junto con la historia, en la misma transacción.
- **Fuera de alcance**, porque no lo piden los criterios o tienen su propia historia:
  - **Editar el título o el responsable** y **eliminar** una tarea.
  - **Registrar horas trabajadas** en una tarea (HU-24).
  - Qué pasa con las tareas de una historia **Hecha** (lo define HU-08 con los estados de la historia).

## 6. Casos límite

Historia de ejemplo: HU-1 "Iniciar sesión con Google". Integrantes activos: Carolina y Mariano.

| Caso | Resultado |
|---|---|
| Primera tarea "Pantalla de login", Carolina, "3" | Se agrega, pendiente; horas de la historia 3 |
| Segunda tarea "Guardar la sesión", Mariano, "2,5" | Se agrega al final; horas de la historia 5,5 |
| Título de solo espacios | Error: título obligatorio |
| Título de 120 caracteres / de 121 | Se agrega / error de largo |
| Título " Pantalla de login " | Se guarda "Pantalla de login" |
| Responsable dado de baja | Error: el responsable tiene que ser un integrante activo |
| Responsable de otro proyecto o que no existe | Error: el responsable tiene que ser un integrante activo |
| Sin responsable | Error: el responsable es obligatorio |
| Horas "0", "-1", "abc" o vacías | Error: las horas estimadas deben ser mayores a 0 |
| Cambiar de "3" a "4" las horas de una tarea pendiente | Quedan 4; horas de la historia 6,5 |
| Cambiar las horas de una tarea completada | Error: la tarea está completada y no cambia su estimación |
| Completar una tarea | Queda completada; las horas de la historia no cambian |
| Completar una que ya lo está | Queda completada, sin error |
| Volver a pendiente una completada y cambiarle las horas | Se puede |
| Valor de marca distinto de `1` y `0` | Error: valor inválido |
| Dos historias con tareas | Cada una muestra y suma solo las suyas |
| Proyecto sin integrantes activos | La página avisa que hay que registrar un integrante antes de agregar tareas |
| Tarea de otra historia en la URL | Página "no encontrado" (404) |
| Proyecto o historia que no existe | Página "no encontrado" (404) |

## 7. Condiciones de error

| Situación | Campo | Error (Go) | Mensaje al usuario |
|---|---|---|---|
| Título vacío o solo espacios | titulo | `ErrTituloTareaObligatorio` | "el título de la tarea es obligatorio" |
| Título de más de 120 caracteres | titulo | `ErrTituloTareaLargo` | "el título de la tarea no puede superar los 120 caracteres" |
| Sin responsable | responsable | `ErrResponsableObligatorio` | "el responsable es obligatorio" |
| Responsable dado de baja, de otro proyecto o inexistente | responsable | `ErrResponsableInvalido` | "el responsable tiene que ser un integrante activo del proyecto" |
| Horas en 0, negativas o que no son un número | horas | `ErrHorasInvalidas` (HU-23) | "las horas estimadas deben ser mayores a 0" |
| Cambiar las horas de una tarea completada | — | `ErrTareaCompletada` | "la tarea está completada: no se puede cambiar su estimación" (aviso en la página, 409) |
| Valor de marca que no es `1` ni `0` | — | (lo controla el handler, antes del caso de uso) | Página 400 "Valor inválido." |
| El proyecto no existe | — | `project.ErrNoEncontrado` | Página 404 "No encontramos ese proyecto" |
| La historia no existe en el proyecto | — | `backlog.ErrHistoriaNoEncontrada` | Página 404 "No encontramos esa historia" |
| La tarea no existe en la historia | — | `effort.ErrTareaNoEncontrada` | Página 404 "No encontramos esa tarea" |

En la pantalla, los errores de una tarea nueva aparecen debajo de cada campo y se conserva lo que el usuario ya había
escrito. Un error al cambiar las horas de una tarea se muestra arriba de la lista.

## 8. Criterios de aceptación

- CA-25.1: Cada tarea tiene título, responsable y horas estimadas.
- CA-25.2: Una tarea pertenece a una sola historia.
- CA-25.3: Una tarea completada no cambia su estimación.

## Trazabilidad

Issue: #3 · Feature: `features/hu-25-descomponer-tareas.feature` · Pasos BDD: `features/hu25_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Dominio | `internal/domain/effort/tareas.go` (`NuevaTarea`, `Tarea.CambiarEstimacion`, `Tarea.Marcar`) y `FormatearHoras` en `horas.go` | `internal/domain/effort/tareas_test.go`, `horas_test.go` |
| Repositorio en memoria | `internal/store/memory/tareas.go` | `internal/store/memory/tareas_test.go` |
| Repositorio PostgreSQL | `internal/store/postgres/tareas.go` | `internal/store/postgres/tareas_test.go` |
| Caso de uso | `internal/service/tareas.go` (`Ver`, `Agregar`, `CambiarEstimacion`, `Marcar`) | `internal/service/tareas_test.go` |
| Pantalla | `internal/server/tareas.go`, `web/templates/paginas/tareas.html` y el enlace "Tareas" de `historia.html` | `internal/server/tareas_test.go` |
| Aplicación | `app/app.go` (elige memoria o PostgreSQL según `DATABASE_URL`) | `app/app_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-25.1 Título, responsable y horas | Agregar la primera tarea de la historia · Agregar varias tareas · Datos inválidos de una tarea · Largo del título de la tarea · Un integrante dado de baja no puede ser responsable | `TestNuevaTarea_*`, `TestAgregarTarea_*`, `TestVerTareas_TraeLasTareasLaSumaYLosIntegrantes`, `TestPaginaDeTareas_*`, `TestTareas_GuardarAsignaIDsYListarPorHistoriaEnOrden`, `TestTareas_LaBaseRechazaDatosInvalidos` |
| CA-25.2 Una sola historia | Cada historia tiene sus propias tareas · Una tarea no se cambia desde otra historia | `TestTareas_BuscarPorID`, `TestTareas_Actualizar`, `TestTareas_DeOtraHistoria_DevuelveNoEncontrada`, `TestTareas_ProyectoHistoriaOTareaInexistente_Responde404` |
| CA-25.3 Una completada no cambia su estimación | Cambiar la estimación de una tarea pendiente · Una tarea completada no cambia su estimación · Completar una tarea y volver a pendiente para corregir las horas | `TestCambiarEstimacion_*`, `TestMarcar_CompletadaOPendienteDeFormaExplicita`, `TestMarcarTarea_*`, `TestCambiarHoras_*` |
| Horas de la historia (CA-23.2, RN8) | Agregar varias tareas | `TestVerTareas_TraeLasTareasLaSumaYLosIntegrantes`, `TestAgregarTarea_VuelveALaPaginaConLaTareaYLaSuma`, `TestFormatearHoras_*` |
| Se guarda en la base (TEC-03) | — | `TestTareas_*` de PostgreSQL, `TestNewHandlerConBase_ConURLLasTareasSobrevivenAUnReinicio` |
