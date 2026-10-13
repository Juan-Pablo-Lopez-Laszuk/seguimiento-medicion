# SPEC-07 · Gestionar criterios de aceptación

**Historia:** HU-07 (#20) · **Responsable:** Piastrellini, Mariano · **Sprint:** 2 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md` y la tabla CriterioAceptacion de
> `docs/modelo-datos.md`.

## 1. Objetivo

Que el Product Builder pueda cargar los criterios de aceptación de una historia del backlog y marcarlos como cumplidos
a medida que se verifican, viendo cuántos faltan (por ejemplo 2/3). Es el camino crítico del Sprint 2: HU-12 solo
asigna al sprint historias con criterios (CA-12.1) y HU-08 solo deja pasar una historia a Hecho cuando todos están
cumplidos (CA-08.2). Esta historia crea los criterios y expone esas dos preguntas; las reglas de HU-12 y de HU-08
las aplican ellas.

## 2. Entradas

**Agregar un criterio**

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Historia | número de la historia (HU-N, viene en la URL) | Sí | Tiene que existir en ese proyecto |
| Descripción | texto | Sí | Hasta 300 caracteres (después de quitar espacios al inicio y al final) |

**Marcar un criterio**

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto, historia y criterio | ids (vienen en la URL) | Sí | El criterio tiene que ser de esa historia |
| Cumplido | `1` (cumplido) o `0` (no cumplido) | Sí | Cualquier otro valor es un error |

El estado inicial de un criterio **no se carga**: nace no cumplido (RN2).

## 3. Salidas esperadas

- **Al agregar:** el criterio queda guardado con id, historia, descripción y **no cumplido**. Se vuelve a la página de
  la historia con el mensaje **"Criterio agregado"** y el criterio nuevo al final de la lista.
- **Al marcar:** el criterio queda cumplido o no cumplido y se vuelve a la página de la historia con el mensaje
  **"Criterio marcado como cumplido"** o **"Criterio marcado como no cumplido"**.
- **Página de la historia** (`/proyectos/{id}/backlog/{numero}`): título, descripción, prioridad, estado y story
  points de la historia; sus criterios en orden de carga, cada uno con su estado y el botón para cambiarlo; el
  formulario para agregar uno nuevo; y el **progreso** ("2/3 criterios"). Sin criterios dice "Sin criterios todavía" y
  avisa que la historia necesita al menos uno para entrar a un sprint.
- **Lista del backlog:** cada historia muestra su progreso ("2/3", o "Sin criterios") y enlaza a su página.

## 4. Reglas de negocio

- RN1: La descripción se guarda **sin los espacios del inicio y del final**; si queda vacía, es obligatoria. El largo
  se cuenta en caracteres, no en bytes.
- RN2: Todo criterio nace **no cumplido** (CA-07.2).
- RN3: Un criterio se marca cumplido o no cumplido **de forma explícita** (el pedido dice a cuál de los dos estados
  pasa), no se "alterna". Así un doble clic o un pedido repetido no lo deja al revés de lo que se quería. Marcar un
  criterio que ya está en ese estado no es un error: queda igual.
- RN4: El **progreso** de una historia es la cantidad de criterios cumplidos sobre el total, y se muestra como
  `cumplidos/total` (CA-07.3). Una historia sin criterios tiene progreso `0/0`.
- RN5: Una historia **tiene criterios** si su total es al menos 1. Esa es la regla de CA-07.1: HU-12 la usa para no
  asignar al sprint una historia sin criterios. Esta historia solo la expone y no asigna nada a ningún sprint.
- RN6: Una historia está **completa** si tiene criterios y todos están cumplidos. Con `0/0` **no** está completa: que
  no haya nada que cumplir no es lo mismo que haberlo cumplido. HU-08 usa esta regla para pasar a Hecho (CA-08.2).
- RN7: Una historia **Hecha no admite cambios en sus criterios**: no se agregan ni se desmarcan, porque eso rompería
  que una historia Hecha tiene todos sus criterios cumplidos (RN6). Es la misma idea que CA-06.2, que protege a las
  historias Hechas.
- RN8: Los criterios de una historia no se mezclan con los de otra: se muestran, se cuentan y se marcan solo dentro de
  su historia. Un criterio con el id de otra historia es un "no encontrado".
- RN9: Los criterios se muestran **en el orden en que se cargaron**.

## 5. Restricciones

- Las reglas de un criterio (RN1, RN2, RN7) y las del progreso (RN4 a RN6) son **funciones puras** en
  `internal/domain/backlog`: reciben los datos y la historia, y devuelven el criterio o el error. No usan base de
  datos ni HTTP, así se prueban con TDD.
- El caso de uso en `internal/service` busca la historia y sus criterios en los repositorios, llama al dominio y
  guarda. Los repositorios son interfaces: `internal/store/memory` en los tests y `internal/store/postgres` en la
  aplicación. **Cambios en los repositorios:** el de historias suma `BuscarPorNumero`, y hay uno nuevo de criterios
  (`ListarPorHistoria`, `ListarPorProyecto` para calcular el progreso de toda la lista de una vez, `Guardar`,
  `BuscarPorID` y `Actualizar`).
- **Sin migración nueva:** la tabla `criterios_aceptacion` ya existe desde TEC-03 (`id`, `historia_id`, `descripcion`,
  `cumplido`), así que no hay nada que aplicar en producción.
- Marcar y agregar se piden con un formulario `POST` (no con un enlace `GET`) y se vuelve a la página con una
  redirección 303, como en el resto de la aplicación. Las pantallas usan los componentes de TEC-07 (`exito`, `aviso`,
  `error`, `errorCampo`).
- Los criterios no tienen `ON DELETE CASCADE` (decisión de TEC-03): qué pasa con ellos cuando se elimina una historia
  lo define HU-06.
- **Fuera de alcance**, porque tienen su propia historia:
  - **Editar o eliminar** un criterio. La historia pide cargar y marcar; si hace falta corregir un texto, se decide
    junto con HU-06 (editar y eliminar ítem).
  - **Cambiar el estado** de la historia (HU-08) y **asignarla a un sprint** (HU-12): HU-07 solo les da las preguntas
    "¿tiene criterios?" y "¿están todos cumplidos?".
  - **Reordenar** los criterios.

## 6. Casos límite

Historia de ejemplo: HU-1 "Iniciar sesión con Google", Pendiente.

| Caso | Resultado |
|---|---|
| Primer criterio de la historia | Se agrega, no cumplido; progreso `0/1` |
| Segundo y tercer criterio | Se agregan al final, en orden; progreso `0/3` |
| Descripción de solo espacios | Error: descripción obligatoria |
| Descripción de 300 caracteres / de 301 | Se agrega / error de largo |
| Descripción " Entra con su cuenta " | Se guarda "Entra con su cuenta" |
| Marcar cumplido un criterio pendiente | Queda cumplido; progreso `1/3` |
| Marcar cumplido uno que ya lo está | Queda cumplido, sin error |
| Marcar no cumplido uno cumplido | Queda no cumplido |
| Valor de marca distinto de `1` y `0` | Error: valor inválido |
| Historia sin criterios | No tiene criterios (RN5) y no está completa (RN6) |
| Historia con `0/2`, `1/2` | Tiene criterios; no está completa |
| Historia con `2/2` | Tiene criterios y está completa |
| Historia Hecha: agregar un criterio | Error: la historia está Hecha |
| Historia Hecha: marcar un criterio como no cumplido | Error: la historia está Hecha |
| Historia En sprint o En progreso | Se pueden agregar y marcar criterios |
| Dos historias del mismo proyecto con criterios | Cada una muestra y cuenta solo los suyos |
| Criterio de otra historia en la URL | Página "no encontrado" (404) |
| Proyecto o historia que no existe | Página "no encontrado" (404) |

## 7. Condiciones de error

| Situación | Campo | Error (Go) | Mensaje al usuario |
|---|---|---|---|
| Descripción vacía o solo espacios | descripcion | `ErrDescripcionCriterioObligatoria` | "la descripción del criterio es obligatoria" |
| Descripción de más de 300 caracteres | descripcion | `ErrDescripcionCriterioLarga` | "la descripción del criterio no puede superar los 300 caracteres" |
| Historia Hecha (agregar o desmarcar) | — | `ErrHistoriaHecha` | "la historia está Hecha: no se pueden cambiar sus criterios" (aviso en la página, 409) |
| Valor de marca que no es `1` ni `0` | — | `ErrValorDeMarcaInvalido` | Página 400 "Valor inválido" |
| El proyecto no existe | — | `project.ErrNoEncontrado` | Página 404 "No encontramos ese proyecto" |
| La historia no existe en el proyecto | — | `backlog.ErrHistoriaNoEncontrada` | Página 404 "No encontramos esa historia" |
| El criterio no existe en la historia | — | `backlog.ErrCriterioNoEncontrado` | Página 404 "No encontramos ese criterio" |

En la pantalla, el error de la descripción aparece debajo del campo y se conserva lo que el usuario ya había escrito.

## 8. Criterios de aceptación

- CA-07.1: Una historia necesita al menos un criterio de aceptación para entrar a un sprint.
- CA-07.2: Cada criterio se marca cumplido o no cumplido.
- CA-07.3: Se muestra el progreso (por ejemplo 2/3).

## Trazabilidad

Issue: #20 · Feature: `features/hu-07-criterios-aceptacion.feature` · Pasos BDD: `features/hu07_steps_test.go`

Se completa al terminar la implementación.
