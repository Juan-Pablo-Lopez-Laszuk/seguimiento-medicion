# SPEC-06 · Editar y eliminar ítem

**Historia:** HU-06 (#19) · **Responsable:** Piastrellini, Mariano · **Sprint:** 2 · **Estado:** En desarrollo

> Especificación escrita antes del código (SDD). Usa la tabla Historia de `docs/modelo-datos.md`, las reglas de
> validación de HU-05 (RN1 a RN3 y RN7), los estados de HU-08 y la nota de TEC-03 (las claves foráneas no tienen
> `ON DELETE CASCADE`: qué pasa con los criterios y las tareas al eliminar una historia lo define esta historia).

## 1. Objetivo

Que el Product Builder pueda corregir una historia del backlog (título, descripción y prioridad) y eliminar la que ya no
hace falta, para mantener el backlog ordenado. Editar nunca cambia el identificador (HU-N), y eliminar tiene tres
cuidados: no se pierde lo que ya es historia del proyecto (una Hecha o la de un sprint cerrado), siempre pide
confirmación y no deja criterios ni tareas huérfanos.

## 2. Entradas

**Editar**

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Historia | número de la historia (HU-N, viene en la URL) | Sí | Tiene que existir en ese proyecto |
| Título | texto | Sí | Hasta 120 caracteres (después de quitar espacios al inicio y al final) |
| Descripción | texto | No | Hasta 2000 caracteres (después de quitar espacios al inicio y al final) |
| Prioridad | opción | Sí | Alta, Media o Baja |

Los story points y el estado **no se editan acá**: se cambian con HU-16 y HU-08.

**Eliminar:** el proyecto y la historia (en la URL) y la **confirmación** (un `POST` aparte de la pantalla que la pide).

## 3. Salidas esperadas

- **Editar, si los datos valen:** la historia queda con los datos nuevos y **el mismo identificador**. Se vuelve a su
  página con el mensaje **"Historia actualizada."**
- **Editar, si no valen:** la historia **no cambia** y cada error aparece debajo de su campo, con lo que se escribió.
- **Eliminar, primera pantalla** (`/proyectos/{id}/backlog/{numero}/eliminar`): pregunta **"¿Eliminar HU-N · título?"**,
  dice que **se eliminan también sus criterios de aceptación y sus tareas** (con la cantidad de cada uno), que no se
  puede deshacer, y ofrece **Eliminar historia** y **Cancelar**. Mostrarla **no elimina nada**.
- **Eliminar, si se confirma y se puede:** la historia, sus criterios y sus tareas dejan de existir. Se vuelve al
  backlog con el mensaje **"Historia HU-N eliminada."**
- **Eliminar, si no se puede:** la pantalla de confirmación explica el motivo y no ofrece eliminar; un `POST` directo
  responde 409 con el mismo motivo y no borra nada.
- **Página de la historia** (`/proyectos/{id}/backlog/{numero}`): suma los enlaces **Editar** y **Eliminar**.
- **Editar** abre `/proyectos/{id}/backlog/{numero}/editar`, el formulario de la historia con sus datos actuales.

## 4. Reglas de negocio

- RN1: Se editan **título, descripción y prioridad**, con las mismas reglas que al crear (HU-05): el título y la
  descripción se guardan sin los espacios del inicio y del final; el título es obligatorio y llega a 120 caracteres; la
  descripción llega a 2000; la prioridad es Alta, Media o Baja. Si hay varios errores, se informan **todos juntos**.
- RN2: **Editar no cambia el identificador** (CA-06.1). Tampoco toca el estado, los story points, los criterios, las
  tareas ni el sprint de la historia.
- RN3: Se puede editar una historia **en cualquier estado**, también una Hecha: corregir un título o una descripción no
  mueve ninguna métrica (a diferencia de re-estimar, que HU-16 no permite en una Hecha).
- RN4: Guardar sin haber cambiado nada vale: la historia queda igual, sin error.
- RN5: **Una historia Hecha no se elimina** (CA-06.2).
- RN6: **Una historia de un sprint cerrado no se elimina** (CA-06.2). Es "de un sprint cerrado" si ese sprint es su
  sprint actual (`historias.sprint_id`) o aparece en la foto de cierre de ese sprint (`sprint_historias`), aunque no se
  haya terminado. Las pruebas de PostgreSQL arman ese estado a mano porque asignar historias a un sprint (HU-12) y
  cerrar el sprint (HU-14) todavía no están en `main`.
- RN7: Una historia **Pendiente, En sprint o En progreso se puede eliminar**, siempre que no sea de un sprint cerrado.
  El criterio no lo prohíbe; la confirmación avisa lo que se pierde.
- RN8: **Eliminar pide confirmación** (CA-06.3). La primera pantalla es un `GET` y no borra nada; solo el `POST` de esa
  pantalla elimina. La pantalla dice cuántos criterios y cuántas tareas se eliminan con la historia.
- RN9: Al eliminar una historia se eliminan **también sus criterios de aceptación y sus tareas**, en **una sola
  operación**: o se borra todo o no se borra nada. No se pueden eliminar de a uno (HU-07 y HU-25 los dejaron fuera de
  alcance), así que bloquear la eliminación de una historia con criterios o tareas la dejaría imposible de eliminar.
- RN10: Si la base rechaza el borrado porque la historia tiene **otros registros asociados** (sesiones de Planning
  Poker, esfuerzo registrado o defectos: funcionalidades que todavía no existen), no se elimina nada, ni siquiera sus
  criterios y tareas, y se informa que tiene registros asociados.
- RN11: **Numeración.** Las demás historias **conservan su número**. La siguiente historia toma el número más alto
  existente más uno (HU-05, RN5): un hueco en el medio no se reutiliza, pero si se elimina la **última**, su número
  queda libre y lo toma la próxima historia. Esto corrige la RN5 de HU-05, que decía que un número nunca se reutiliza.

## 5. Restricciones

- Las reglas son **funciones puras** en `internal/domain/backlog`: `Historia.Editar` (RN1 a RN4) y
  `Historia.PuedeEliminarse` (RN5 a RN7, recibe si la historia es de un sprint cerrado). No usan base de datos ni HTTP, así
  se prueban con TDD. Reusan la validación de HU-05.
- El caso de uso en `internal/service` (`EdicionHistorias`) busca el proyecto y la historia, consulta si es de un
  sprint cerrado, cuenta sus criterios y tareas, llama a las reglas y guarda o elimina. Los repositorios son interfaces.
- **Cambios en el repositorio de historias** (memoria y PostgreSQL): `ActualizarDatos` (título, descripción y
  prioridad), `EnSprintCerrado` y `Eliminar`.
  - En **PostgreSQL**, `Eliminar` borra los criterios, las tareas y la historia en **una transacción**; una violación de
    clave foránea (`23503`) vuelve atrás todo y devuelve `backlog.ErrHistoriaConRegistros` (RN10).
  - En **memoria**, `memory.NuevoHistorias` acepta los repositorios que dependen de la historia (criterios y tareas) y
    les pide que borren lo suyo; sin argumentos funciona igual que hasta ahora.
  - `EnSprintCerrado` en memoria devuelve `false` mientras HU-12 no esté en `main`: hoy en memoria no se puede
    asignar una historia a un sprint. Quien integre HU-12 con esta historia tiene que hacer que consulte el estado del
    sprint (queda anotado en el PR).
- **Sin migración nueva:** no se agrega `ON DELETE CASCADE` (decisión de TEC-03); la cascada la hace el repositorio,
  explícita y atómica, y no hay nada que aplicar en producción.
- Editar y eliminar se piden con formularios `POST` y redirecciones 303, como en el resto de la aplicación. Las
  pantallas usan los componentes de TEC-07 (`exito`, `aviso`, `error`, `errorCampo`).
- **Confirmación como pantalla propia, no como ventana modal.** Tiene que decir cuántos criterios y tareas se
  eliminan y por qué no se puede eliminar; esos datos no los tiene la página de la historia, que se arma en cuatro
  lugares. Una pantalla propia también deja la confirmación **verificable en el servidor** (RN8).
- **Fuera de alcance**, porque tienen su propia historia o no lo pide el criterio:
  - Editar los **story points** (HU-16) y el **estado** (HU-08), y **reordenar** el backlog (HU-09).
  - **Editar o eliminar un criterio** o una **tarea** suelta (HU-07 y HU-25).
  - **Deshacer** una eliminación o una papelera, y registrar **quién** eliminó.
  - Eliminar desde la lista del backlog: se elimina desde la página de la historia.

## 6. Casos límite

Historia de ejemplo: HU-2 "Registro con email", Alta, con 2 criterios y 1 tarea. El proyecto tiene HU-1, HU-2 y HU-3.

| Caso | Resultado |
|---|---|
| Editar título, descripción y prioridad con datos válidos | Cambian; sigue siendo HU-2 |
| Editar una historia con 5 story points, En progreso, con criterios y tareas | Los story points, el estado, los criterios y las tareas quedan igual |
| Guardar sin cambiar nada | Vale; queda igual |
| Título de solo espacios o vacío | Error: título obligatorio; la historia no cambia |
| Título de 120 caracteres / de 121 | Se guarda / error de largo |
| Descripción de 2000 caracteres / de 2001 | Se guarda / error de largo |
| Prioridad "Urgente" | Error: prioridad inválida |
| Título vacío y prioridad inválida a la vez | Los dos errores juntos |
| Editar una historia Hecha | Vale |
| Mostrar la pantalla de eliminar | Dice "2 criterios y 1 tarea"; no elimina |
| Confirmar la eliminación de HU-2 | HU-2 no está más; HU-1 y HU-3 conservan su número; sus criterios y tareas tampoco |
| Crear una historia después de eliminar la HU-2 | Es la HU-4 |
| Eliminar la HU-3 (la última) y crear otra | La nueva es la HU-3 |
| Eliminar una historia Pendiente, En sprint o En progreso | Se elimina |
| Eliminar una historia Hecha | Error: no se puede eliminar (409) |
| Eliminar una historia de un sprint cerrado (aunque no esté Hecha) | Error: no se puede eliminar (409) |
| La base rechaza el borrado por otros registros asociados | Error: tiene registros asociados (409); no se borra nada |
| Confirmar una eliminación de una historia que ya no existe | Página "no encontrado" (404) |
| Proyecto o historia que no existe | Página "no encontrado" (404) |

## 7. Condiciones de error

| Situación | Campo | Error (Go) | Mensaje al usuario |
|---|---|---|---|
| Título vacío | titulo | `ErrTituloObligatorio` (de HU-05) | "el título es obligatorio" |
| Título de más de 120 caracteres | titulo | `ErrTituloLargo` (de HU-05) | "el título no puede superar los 120 caracteres" |
| Descripción de más de 2000 caracteres | descripcion | `ErrDescripcionLarga` (de HU-05) | "la descripción no puede superar los 2000 caracteres" |
| Prioridad que no es Alta, Media ni Baja | prioridad | `ErrPrioridadInvalida` (de HU-05) | "la prioridad tiene que ser Alta, Media o Baja" |
| La historia está Hecha (eliminar) | — | `ErrEliminarHistoriaHecha` | "la historia está Hecha: no se puede eliminar" (409) |
| La historia es de un sprint cerrado (eliminar) | — | `ErrEliminarEnSprintCerrado` | "la historia es de un sprint cerrado: no se puede eliminar" (409) |
| La base rechaza el borrado por otros registros (eliminar) | — | `ErrHistoriaConRegistros` | "la historia tiene registros asociados (Planning Poker, esfuerzo o defectos): no se puede eliminar" (409) |
| El proyecto no existe | — | `project.ErrNoEncontrado` | Página 404 "No encontramos ese proyecto" |
| La historia no existe en el proyecto | — | `backlog.ErrHistoriaNoEncontrada` | Página 404 "No encontramos esa historia" |

En el formulario de editar, los errores de los campos aparecen debajo de cada uno (422).

## 8. Criterios de aceptación

- CA-06.1: Editar no cambia el identificador.
- CA-06.2: No se eliminan historias Hechas ni asignadas a un sprint cerrado.
- CA-06.3: Eliminar pide confirmación.

## Trazabilidad

Issue: #19 · Feature: `features/hu-06-editar-eliminar-item.feature` · Pasos BDD: `features/hu06_steps_test.go`

_Se completa cuando la historia está implementada._
