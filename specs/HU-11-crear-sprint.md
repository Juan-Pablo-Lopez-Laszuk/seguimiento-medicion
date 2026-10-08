# SPEC-11 · Crear sprint con Sprint Goal

**Historia:** HU-11 (#43) · **Responsable:** López, Juan Pablo · **Sprint:** 1 · **Estado:** Revisada

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md` y la tabla Sprint de
> `docs/modelo-datos.md`.

## 1. Objetivo

Que el Agile Enabler pueda crear los sprints de un proyecto, con sus fechas y su Sprint Goal, para planificar cada
iteración. Es la base de la épica E3: después se le asignan historias (HU-12), se completan (HU-13) y se cierra
(HU-14), y de ahí salen la velocidad y el resto de las métricas.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Sprint Goal | texto | Sí | Hasta 500 caracteres (después de quitar espacios al inicio y al final) |
| Fecha de inicio | fecha (`AAAA-MM-DD`) | Sí | Fecha válida, dentro del proyecto y posterior al último sprint |
| Fecha de finalización | fecha (`AAAA-MM-DD`) | Sí | Fecha válida, posterior a la de inicio y dentro del proyecto |

El número del sprint **no se carga**: lo asigna el sistema (RN5).

## 3. Salidas esperadas

- El sprint queda guardado con: id, proyecto, **número correlativo**, Sprint Goal, fechas y **estado Planificado**.
- En la interfaz se vuelve a la lista de sprints del proyecto con el mensaje **"Sprint N creado"** y el sprint nuevo
  visible.

## 4. Reglas de negocio

- RN1: El Sprint Goal se guarda **sin los espacios del inicio y del final**; si queda vacío, es obligatorio
  (CA-11.1). El largo se cuenta en caracteres, no en bytes.
- RN2: La fecha de finalización tiene que ser **estrictamente posterior** a la de inicio (igual que en HU-01).
- RN3: El sprint tiene que quedar **dentro del proyecto**: el inicio no puede ser anterior al inicio del proyecto y el
  fin no puede ser posterior al fin del proyecto. Los bordes valen: un sprint puede empezar el mismo día que el
  proyecto y terminar el mismo día que él (CA-11.2).
- RN4: Los sprints se crean **en orden**: el nuevo tiene que empezar **después del día en que termina el último
  sprint** del proyecto. Así no se superponen (CA-11.2) y el número siempre sigue el orden de las fechas, que es lo
  que necesitan la velocidad (HU-31 usa "los últimos sprints") y la lista de HU-15 (ordenada por número). Puede
  haber días libres entre un sprint y el siguiente.
- RN5: El número es **correlativo dentro del proyecto**: el primer sprint es el 1 y cada nuevo es el número más alto
  más uno. Cada proyecto tiene su propia numeración (CA-11.3).
- RN6: Todo sprint nuevo nace en estado **Planificado** (CA-11.3). Pasa a Activo y a Cerrado en otras historias.
- RN7: Si hay varios datos inválidos, se informan **todos juntos**, cada uno con su campo (como en HU-01).

## 5. Restricciones

- Las reglas RN1 a RN7 son una **función pura** en `internal/domain/sprint`: recibe los datos, el proyecto y los
  sprints que ya tiene, y devuelve el sprint o los errores. No usa base de datos ni HTTP, así se prueba con TDD.
- El caso de uso en `internal/service` busca el proyecto y sus sprints en los repositorios, llama al dominio y guarda.
  Los repositorios son interfaces: `internal/store/memory` en los tests y `internal/store/postgres` cuando esté
  TEC-03. **Para TEC-03:** el repositorio de proyectos suma `BuscarPorID` y hace falta uno de sprints
  (`ListarPorProyecto` y `Guardar`).
- `ErroresValidacion` pasa de `internal/domain/project` a `internal/domain` para que la usen proyecto y sprint
  (refactor sin cambio de comportamiento de HU-01).
- En la base, la restricción única `(proyecto_id, numero)` de `docs/modelo-datos.md` evita dos sprints con el mismo
  número si dos personas crean un sprint al mismo tiempo.
- Las fechas se manejan sin hora (`time.Time` a medianoche UTC), como en HU-01.
- Hasta que exista la vista del proyecto (HU-04, Sprint 2), a los sprints se llega desde la lista de proyectos.

## 6. Casos límite

Proyecto de ejemplo: del 05/10/2026 al 01/11/2026.

| Caso | Resultado |
|---|---|
| Primer sprint del proyecto | Se crea con el número 1 |
| Sprint Goal de solo espacios | Error: Sprint Goal obligatorio |
| Sprint Goal de 500 caracteres / de 501 | Se crea / error de largo |
| Inicio el mismo día que el proyecto (05/10) | Se crea |
| Inicio un día antes que el proyecto (04/10) | Error de fecha de inicio fuera del proyecto |
| Fin el mismo día que el proyecto (01/11) | Se crea |
| Fin un día después que el proyecto (02/11) | Error de fecha de fin fuera del proyecto |
| Fin igual al inicio | Error de fechas |
| Ya existe el Sprint 1 (05/10 al 11/10) y el nuevo empieza el 11/10 | Error: se superpone |
| Ya existe el Sprint 1 (05/10 al 11/10) y el nuevo empieza el 12/10 | Se crea con el número 2 |
| Ya existe el Sprint 2 (12/10 al 18/10) y el nuevo es del 05/10 al 08/10 (antes del último) | Error: tiene que empezar después del último sprint |
| Otro proyecto ya tiene 3 sprints | El primer sprint de este proyecto igual es el 1 |
| El proyecto no existe | Página "no encontrado" (404) |

## 7. Condiciones de error

| Situación | Campo | Error (Go) | Mensaje al usuario |
|---|---|---|---|
| Sprint Goal vacío o solo espacios | objetivo | `ErrObjetivoObligatorio` | "el Sprint Goal es obligatorio" |
| Sprint Goal de más de 500 caracteres | objetivo | `ErrObjetivoLargo` | "el Sprint Goal no puede superar los 500 caracteres" |
| Falta la fecha de inicio o no es una fecha | fecha_inicio | `ErrFechaInicio` | "la fecha de inicio es obligatoria y debe ser válida" |
| Falta la fecha de fin o no es una fecha | fecha_fin | `ErrFechaFin` | "la fecha de finalización es obligatoria y debe ser válida" |
| Fecha de fin igual o anterior a la de inicio | fecha_fin | `ErrFechasInvalidas` | "la fecha de finalización debe ser posterior a la de inicio" |
| Inicio antes del proyecto o después de su fin | fecha_inicio | `ErrFueraDelProyecto` | "la fecha tiene que estar dentro del proyecto (del 05/10/2026 al 01/11/2026)" |
| Fin después del fin del proyecto | fecha_fin | `ErrFueraDelProyecto` | "la fecha tiene que estar dentro del proyecto (del 05/10/2026 al 01/11/2026)" |
| Inicio igual o anterior al fin del último sprint | fecha_inicio | `ErrSuperpuesto` | "el sprint tiene que empezar después del último (el Sprint 1 termina el 11/10/2026)" |
| El proyecto no existe | — | `project.ErrNoEncontrado` | Página 404 "No encontramos ese proyecto" |

Los mensajes con fechas se arman envolviendo el error (`fmt.Errorf("%w ...")`), así los tests comparan con
`errors.Is` y el usuario ve las fechas concretas. En la pantalla, cada mensaje aparece debajo de su campo y se
conservan los datos que el usuario ya había cargado.

## 8. Criterios de aceptación

- CA-11.1: El Sprint Goal es obligatorio.
- CA-11.2: Las fechas quedan dentro del rango del proyecto y no se superponen con otro sprint.
- CA-11.3: El número de sprint es correlativo y el estado inicial es Planificado.

## Trazabilidad

Issue: #43 · Feature: `features/hu-11-crear-sprint.feature` · Pasos BDD: `features/hu11_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Dominio | `internal/domain/sprint/sprint.go` | `internal/domain/sprint/sprint_test.go` |
| Repositorio en memoria | `internal/store/memory/sprints.go` | `internal/store/memory/sprints_test.go` |
| Caso de uso | `internal/service/sprints.go` | `internal/service/sprints_test.go` |
| Pantalla | `internal/server/sprints.go`, `web/templates/paginas/sprints.html`, `sprint_nuevo.html` | `internal/server/sprints_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-11.1 Sprint Goal obligatorio | Crear el primer sprint · Sprint Goal vacío | se completa con el código |
| CA-11.2 Dentro del proyecto y sin superponerse | Fechas en el borde del proyecto · Se superpone con el último sprint | se completa con el código |
| CA-11.3 Número correlativo y estado Planificado | Crear el primer sprint · El siguiente sprint lleva el número 2 · Cada proyecto numera sus sprints | se completa con el código |
