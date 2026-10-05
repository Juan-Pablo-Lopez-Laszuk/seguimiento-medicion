# SPEC-01 · Crear proyecto

**Historia:** HU-01 (#39) · **Responsable:** López, Juan Pablo · **Sprint:** 1 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md` y la tabla Proyecto de
> `docs/modelo-datos.md`.

## 1. Objetivo

Que el Agile Enabler pueda dar de alta un proyecto con su nombre, descripción y fechas, para empezar a cargar su backlog, sus integrantes y sus sprints. Es la primera historia: todas las demás cuelgan de un proyecto.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Nombre | texto | Sí | 3 a 100 caracteres (después de quitar espacios al inicio y al final); único |
| Descripción | texto | No | Hasta 1000 caracteres |
| Fecha de inicio | fecha (`AAAA-MM-DD`) | Sí | Fecha válida |
| Fecha de finalización | fecha (`AAAA-MM-DD`) | Sí | Fecha válida, posterior a la de inicio |

## 3. Salidas esperadas

- El proyecto queda guardado con: id, nombre, descripción, fechas, **estado Planificado** y fecha y hora de creación.
- En la interfaz se vuelve a la lista de proyectos con el mensaje **"Proyecto creado"** y el proyecto nuevo visible.

## 4. Reglas de negocio

- RN1: El nombre se guarda **sin los espacios del inicio y del final** (" Metrics " se guarda como "Metrics").
- RN2: El largo del nombre se cuenta en **caracteres**, no en bytes: "Gestión" tiene 7 caracteres aunque la "ó" ocupe
  2 bytes.
- RN3: No puede haber dos proyectos con el mismo nombre **sin distinguir mayúsculas de minúsculas** ("Metrics" y
  "metrics" se consideran iguales).
- RN4: La fecha de finalización tiene que ser **estrictamente posterior** a la de inicio (el mismo día no vale).
- RN5: La fecha de inicio **puede estar en el pasado**: se puede cargar un proyecto que ya empezó.
- RN6: Todo proyecto nuevo nace en estado **Planificado**. El estado después lo calcula HU-04 a partir de las fechas
  y los sprints; acá no se elige.
- RN7: Si hay varios datos inválidos, se informan **todos juntos**, cada uno con su campo, para que el usuario los
  corrija de una sola vez.

## 5. Restricciones

- Las validaciones de los datos (RN1, RN2, RN4, RN6) son **funciones puras** en `internal/domain/project`: no usan
  base de datos ni HTTP, así se prueban con TDD.
- La unicidad del nombre (RN3) necesita consultar los proyectos existentes: la controla el caso de uso en
  `internal/service`, a través del repositorio (`internal/store/memory` en los tests, `internal/store/postgres` en
  producción). En la base, el índice único va sobre `lower(nombre)` (avisar a TEC-03).
- Las fechas se manejan sin hora (`time.Time` a medianoche UTC) para que la comparación no dependa de la zona horaria.
- Hasta que exista el login (TEC-05, Sprint 2) cualquier visitante puede crear proyectos; después, solo un usuario
  logueado, que queda como Agile Enabler del proyecto (HU-03).

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Nombre de 2 caracteres | Error de largo |
| Nombre de 3 caracteres | Se crea |
| Nombre de 100 caracteres | Se crea |
| Nombre de 101 caracteres | Error de largo |
| Nombre "   ab   " (2 caracteres sin los espacios) | Error de largo |
| Nombre con tildes o ñ ("Año Ñandú") | Se cuentan caracteres, no bytes |
| Descripción vacía | Se crea sin descripción |
| Descripción de 1000 caracteres / de 1001 | Se crea / error |
| Fecha de fin igual a la de inicio | Error de fechas |
| Fecha de fin un día después de la de inicio | Se crea |
| Nombre igual a uno existente con otras mayúsculas | Error de nombre repetido |

## 7. Condiciones de error

| Situación | Campo | Error (Go) | Mensaje al usuario |
|---|---|---|---|
| Nombre vacío o solo espacios | nombre | `ErrNombreObligatorio` | "el nombre es obligatorio" |
| Nombre con menos de 3 o más de 100 caracteres | nombre | `ErrNombreLargo` | "el nombre debe tener entre 3 y 100 caracteres" |
| Ya existe un proyecto con ese nombre | nombre | `ErrNombreRepetido` | "ya existe un proyecto con ese nombre" |
| Descripción de más de 1000 caracteres | descripcion | `ErrDescripcionLarga` | "la descripción no puede superar los 1000 caracteres" |
| Falta la fecha de inicio o no es una fecha | fecha_inicio | `ErrFechaInicio` | "la fecha de inicio es obligatoria y debe ser válida" |
| Falta la fecha de fin o no es una fecha | fecha_fin | `ErrFechaFin` | "la fecha de finalización es obligatoria y debe ser válida" |
| Fecha de fin igual o anterior a la de inicio | fecha_fin | `ErrFechasInvalidas` | "la fecha de finalización debe ser posterior a la de inicio" |

Los errores se devuelven agrupados por campo (RN7). En la pantalla, cada mensaje aparece debajo de su campo y se
conservan los datos que el usuario ya había cargado.

## 8. Criterios de aceptación

- CA-01.1: El nombre es obligatorio, tiene entre 3 y 100 caracteres y no se repite.
- CA-01.2: La fecha de finalización debe ser posterior a la de inicio.
- CA-01.3: El proyecto se crea en estado Planificado.
- CA-01.4: Ante un dato inválido se indica qué campo falló.

## Trazabilidad

Issue: #39 · Feature: `features/hu-01-crear-proyecto.feature` · Tests: `internal/domain/project/project_test.go`,
`internal/service/proyectos_test.go` · Código: `internal/domain/project/project.go`, `internal/service/proyectos.go`,
`internal/server` (pantalla) · PR: #
