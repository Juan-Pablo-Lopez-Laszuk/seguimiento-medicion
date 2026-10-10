# SPEC-05 · Crear ítem del Product Backlog

**Historia:** HU-05 (#18) · **Responsable:** Piastrellini, Mariano · **Sprint:** 1 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md` y la tabla Historia de
> `docs/modelo-datos.md`.

## 1. Objetivo

Que el Product Builder pueda agregar historias al Product Backlog de un proyecto, con su título, descripción,
prioridad y, si ya las conoce, sus story points, para registrar el trabajo pendiente. Es la base de la épica E2:
después se editan (HU-06), se les cargan criterios de aceptación (HU-07), cambian de estado (HU-08), se estiman
(HU-16) y se asignan a un sprint (HU-12).

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Título | texto | Sí | Hasta 120 caracteres (después de quitar espacios al inicio y al final) |
| Descripción | texto | No | Hasta 2000 caracteres (después de quitar espacios al inicio y al final) |
| Prioridad | opción | Sí | Alta, Media o Baja |
| Story points | número | No | Vacío (sin estimar) o uno de la escala 0, 1, 2, 3, 5, 8, 13, 21 |

El identificador (HU-1, HU-2...) **no se carga**: lo asigna el sistema (RN5). El estado tampoco: toda historia nace
Pendiente (RN6).

## 3. Salidas esperadas

- La historia queda guardada con: id, proyecto, **número correlativo**, título, descripción, prioridad, **estado
  Pendiente** y story points (o sin estimar).
- En la interfaz se vuelve a la lista del backlog del proyecto con el mensaje **"Historia HU-N creada"** y la
  historia nueva visible, con su identificador **HU-N**.

## 4. Reglas de negocio

- RN1: El título se guarda **sin los espacios del inicio y del final**; si queda vacío, es obligatorio (CA-05.2). El
  largo se cuenta en caracteres, no en bytes.
- RN2: La descripción es opcional. Se guarda sin los espacios del inicio y del final; vacía queda como texto vacío.
- RN3: La prioridad es **Alta, Media o Baja** (CA-05.2). Cualquier otro valor, o ninguno, es un error: no se elige
  una prioridad por el usuario.
- RN4: Los story points son opcionales (CA-05.3). Si se cargan, tienen que ser uno de la **escala acordada**
  (0, 1, 2, 3, 5, 8, 13, 21), la misma que valida la base y que HU-16 usa para estimar. Vacío significa "sin
  estimar", y **no es lo mismo que 0**.
- RN5: El número es **correlativo dentro del proyecto**: la primera historia es la HU-1 y cada nueva es el número
  más alto más uno (CA-05.1). Cada proyecto tiene su propia numeración, y un número nunca se reutiliza.
- RN6: Toda historia nueva nace en estado **Pendiente** (CA-05.3). Los otros estados llegan con HU-08.
- RN7: Si hay varios datos inválidos, se informan **todos juntos**, cada uno con su campo (como en HU-01).

## 5. Restricciones

- Las reglas RN1 a RN7 son una **función pura** en `internal/domain/backlog`: recibe los datos, el proyecto y las
  historias que ya tiene, y devuelve la historia o los errores. No usa base de datos ni HTTP, así se prueba con TDD.
- El caso de uso en `internal/service` busca el proyecto y sus historias en los repositorios, llama al dominio y
  guarda. Los repositorios son interfaces: `internal/store/memory` en los tests y `internal/store/postgres` en la
  aplicación (la tabla `historias` ya existe desde TEC-03).
- En la base, la restricción única `(proyecto_id, numero)` evita dos historias con el mismo número si dos personas
  crean una al mismo tiempo. Igual que en HU-03 y HU-11, hoy ese choque se muestra como error general y no como
  "ya existe"; traducirlo es una mejora pendiente para las tres historias.
- **Fuera de alcance**, porque tienen su propia historia:
  - Los **criterios de aceptación** de la historia (HU-07, Sprint 2). Una historia se crea sin criterios.
  - Las **horas estimadas** (HU-23): la regla de Carolina ya está; guardarlas y mostrarlas sobre la historia se
    agrega encima de esta tabla sin cambiar su estructura.
  - Editar y eliminar (HU-06), cambiar el estado (HU-08), estimar después de creada (HU-16) y el orden manual
    (HU-09): la lista se muestra por número y la columna `orden` queda en su valor por defecto.
- Hasta que exista la vista del proyecto (HU-04, Sprint 2), al backlog se llega desde la lista de proyectos.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Primera historia del proyecto | Se crea como HU-1, Pendiente |
| Título de solo espacios | Error: título obligatorio |
| Título de 120 caracteres / de 121 | Se crea / error de largo |
| Título " Login con Google " | Se guarda "Login con Google" |
| Descripción vacía | Se crea con descripción vacía |
| Descripción de 2000 caracteres / de 2001 | Se crea / error de largo |
| Prioridad Alta, Media o Baja | Se crea |
| Prioridad vacía o "Urgente" | Error de prioridad |
| Story points vacío | Se crea sin estimar (no es 0) |
| Story points 0 y 21 (extremos de la escala) | Se crea |
| Story points 4, 22, -1, "3,5" o "abc" | Error de story points |
| Ya existe la HU-1 | La nueva es la HU-2 |
| Otro proyecto ya tiene 3 historias | La primera historia de este proyecto igual es la HU-1 |
| Varios datos inválidos a la vez | Se informan todos juntos |
| El proyecto no existe | Página "no encontrado" (404) |

## 7. Condiciones de error

| Situación | Campo | Error (Go) | Mensaje al usuario |
|---|---|---|---|
| Título vacío o solo espacios | titulo | `ErrTituloObligatorio` | "el título es obligatorio" |
| Título de más de 120 caracteres | titulo | `ErrTituloLargo` | "el título no puede superar los 120 caracteres" |
| Descripción de más de 2000 caracteres | descripcion | `ErrDescripcionLarga` | "la descripción no puede superar los 2000 caracteres" |
| Prioridad vacía o desconocida | prioridad | `ErrPrioridadInvalida` | "la prioridad tiene que ser Alta, Media o Baja" |
| Story points fuera de la escala o que no son un número | story_points | `ErrStoryPointsInvalidos` | "los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21" |
| El proyecto no existe | — | `project.ErrNoEncontrado` | Página 404 "No encontramos ese proyecto" |

En la pantalla, cada mensaje aparece debajo de su campo y se conservan los datos que el usuario ya había cargado.

## 8. Criterios de aceptación

- CA-05.1: El identificador es autogenerado y correlativo por proyecto (HU-1, HU-2, ...).
- CA-05.2: El título es obligatorio y la prioridad es Alta, Media o Baja.
- CA-05.3: El estado inicial es Pendiente y los story points son opcionales al crear.

## Trazabilidad

Issue: #18 · Feature: `features/hu-05-crear-item-backlog.feature` · Pasos BDD: `features/hu05_steps_test.go`

Se completa al terminar la implementación.
