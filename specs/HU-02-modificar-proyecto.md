# SPEC-02 · Modificar proyecto

**Historia:** HU-02 (#40) · **Responsable:** López, Juan Pablo · **Sprint:** 2 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Reutiliza las reglas de `specs/HU-01-crear-proyecto.md`, la de
> sprints de `specs/HU-11-crear-sprint.md` y el estado calculado de `specs/HU-04-estado-proyecto.md`.

## 1. Objetivo

Que el Agile Enabler pueda corregir o actualizar el nombre, la descripción y las fechas de un proyecto, sin dejar
sprints afuera del proyecto y sin tocar un proyecto que ya terminó.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir y no estar Finalizado |
| Nombre | texto | Sí | Igual que HU-01: 3 a 100 caracteres y único (sin contar al propio proyecto) |
| Descripción | texto | No | Igual que HU-01: hasta 1000 caracteres |
| Fecha de inicio | fecha | Sí | Igual que HU-01 y no posterior al inicio del primer sprint |
| Fecha de finalización | fecha | Sí | Igual que HU-01 y no anterior al fin del último sprint |

## 3. Salidas esperadas

- El proyecto queda guardado con los datos nuevos. No cambian su id, su fecha de creación ni sus sprints e
  integrantes.
- En la interfaz se vuelve a la vista del proyecto (HU-04) con el mensaje **"Proyecto modificado"**.
- El formulario de edición aparece **con los datos actuales ya cargados**.

## 4. Reglas de negocio

- RN1: Aplican **las mismas validaciones que al crear** (HU-01 RN1 a RN7): nombre sin espacios alrededor, de 3 a 100
  caracteres; descripción de hasta 1000; fechas obligatorias y fin posterior al inicio; todos los errores juntos
  (CA-02.1).
- RN2: El nombre no puede repetir el de **otro** proyecto (sin distinguir mayúsculas). Dejar el mismo nombre o cambiarle
  solo las mayúsculas ("software metrics" → "Software Metrics") sí vale.
- RN3: La fecha de fin **no puede quedar antes del fin del último sprint** (CA-02.2). Puede ser el mismo día.
- RN4: La fecha de inicio **no puede quedar después del inicio del primer sprint**. Puede ser el mismo día. No está
  escrita en la historia, pero sale de HU-11 RN3 (todo sprint queda dentro del proyecto): sin esta regla, achicar el
  inicio dejaría al Sprint 1 afuera. **A confirmar en el refinamiento.**
- RN5: Un proyecto **Finalizado** (estado calculado de HU-04 RN1, con "hoy" en la hora de Argentina) **no se puede
  modificar** (CA-02.3): no se muestra el formulario y, si igual llega el pedido, se rechaza.
- RN6: El estado no se edita: se sigue calculando (HU-04).
- RN7: Guardar sin cambiar nada no es un error: vuelve a la vista con el mensaje.

## 5. Restricciones

- `project.Nuevo` de HU-01 se reutiliza para RN1. RN3 a RN5 son funciones puras que reciben el proyecto, sus sprints y
  "hoy" (en `internal/domain/project`).
- El caso de uso busca el proyecto y sus sprints, valida y guarda. El repositorio de proyectos necesita
  `Actualizar(ctx, p)` y que la búsqueda de nombre repetido pueda **excluir al propio proyecto**
  (`ExisteNombre(ctx, nombre, exceptoID)` o una función aparte). TEC-03 no los trajo, así que se agregan con esta
  historia **en los dos repositorios** (memoria y `internal/store/postgres`, con su test contra la base). El índice
  único sobre `lower(nombre)` ya lo cubre en la base.
- Depende de **HU-04** (estado calculado y vista del proyecto): se implementa después.

## 6. Casos límite

Proyecto "Software Metrics" del 05/10/2026 al 01/11/2026, con el Sprint 1 del 05/10 al 11/10 y el Sprint 2 del 12/10 al
18/10. Existe además el proyecto "Otro proyecto".

| Caso | Resultado |
|---|---|
| Cambiar solo la descripción | Se guarda |
| Cambiar el nombre a "software metrics" (el propio, en minúsculas) | Se guarda |
| Cambiar el nombre a "OTRO PROYECTO" | Error de nombre repetido |
| Fin el 18/10 (el mismo día que termina el Sprint 2) | Se guarda |
| Fin el 17/10 | Error: el Sprint 2 termina el 18/10/2026 |
| Inicio el 05/10 (el mismo día que empieza el Sprint 1) | Se guarda |
| Inicio el 06/10 | Error: el Sprint 1 empieza el 05/10/2026 |
| Proyecto sin sprints: cualquier fecha válida | Se guarda |
| Hoy 02/11 (proyecto con sprints y fecha de fin pasada → Finalizado) | No se puede modificar |
| Nombre de 2 caracteres y fin anterior al inicio | Se informan los dos errores juntos |
| El proyecto no existe | Página "no encontrado" (404) |

## 7. Condiciones de error

| Situación | Campo | Error (Go) | Mensaje al usuario |
|---|---|---|---|
| Las de HU-01 | los de HU-01 | los de HU-01 | los de HU-01 |
| Fin antes del fin del último sprint | fecha_fin | `ErrFinAntesDelUltimoSprint` | "la fecha de finalización no puede ser anterior al fin del último sprint (el Sprint 2 termina el 18/10/2026)" |
| Inicio después del inicio del primer sprint | fecha_inicio | `ErrInicioDespuesDelPrimerSprint` | "la fecha de inicio no puede ser posterior al inicio del primer sprint (el Sprint 1 empieza el 05/10/2026)" |
| Proyecto Finalizado | — | `ErrProyectoFinalizado` | "un proyecto Finalizado no se puede modificar" (aviso en la vista, 409) |
| El proyecto no existe | — | `project.ErrNoEncontrado` | Página 404 "No encontramos ese proyecto" |

## 8. Criterios de aceptación

- CA-02.1: Aplican las mismas validaciones que al crear.
- CA-02.2: La fecha de fin no puede quedar antes del fin del último sprint.
- CA-02.3: Un proyecto Finalizado no se puede modificar.

## Trazabilidad

Issue: #40 · Feature: `features/hu-02-modificar-proyecto.feature` · Pasos BDD: `features/hu02_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Dominio | `internal/domain/project/project.go` (`Modificar`) | `internal/domain/project/project_test.go` |
| Repositorio en memoria | `internal/store/memory/proyectos.go` (`Actualizar`, nombre repetido sin contar al propio) | `internal/store/memory/proyectos_test.go` |
| Repositorio PostgreSQL | `internal/store/postgres/proyectos.go` (lo mismo) | `internal/store/postgres/proyectos_test.go` |
| Caso de uso | `internal/service/proyectos.go` (`Modificar`) | `internal/service/proyectos_test.go` |
| Pantalla | `internal/server/proyectos.go`, `web/templates/paginas/proyecto_editar.html` | `internal/server/proyectos_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-02.1 Mismas validaciones que al crear | Modificar la descripción · Nombre propio con otras mayúsculas · Nombre de otro proyecto · Errores juntos | se completa con el código |
| CA-02.2 Fin no antes del último sprint | Fechas contra los sprints | se completa con el código |
| CA-02.3 Finalizado no se modifica | Un proyecto Finalizado no se modifica | se completa con el código |
