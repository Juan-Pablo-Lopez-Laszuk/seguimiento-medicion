# SPEC-09 · Priorizar el backlog

**Historia:** HU-09 (#22) · **Responsable:** Piastrellini, Mariano · **Sprint:** 3 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`. Parte de HU-05 (la prioridad de
> cada historia y su número) y de HU-06 (editar la prioridad). La columna `historias.orden` existe desde TEC-03, "para el
> orden manual dentro de la misma prioridad (HU-09)" (`docs/modelo-datos.md`): **no hace falta una migración**.

## 1. Objetivo

Que el equipo vea el Product Backlog **ordenado por prioridad** (Alta, Media, Baja) y pueda **ordenar a mano** las
historias dentro de una misma prioridad, subiéndolas o bajándolas un lugar, para decidir qué entra en el próximo sprint.
El orden queda guardado y las historias Hechas se muestran aparte.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto (en la URL) | ID | Sí | Tiene que existir |
| Historia (en la URL) | número `HU-N` del proyecto | Sí | Tiene que existir en el proyecto |
| Movimiento | subir o bajar (la acción de la URL: `…/subir` o `…/bajar`) | Sí | Uno de los dos |

## 3. Salidas esperadas

- **La página del backlog** (`/proyectos/{id}/backlog`) muestra dos secciones:
  1. **Pendientes y en curso**: las historias que no están Hechas, ordenadas por prioridad (Alta, Media, Baja) y, dentro de
     cada prioridad, por el orden manual. Cada fila tiene el botón **Subir**, salvo la primera de su prioridad, y el botón
     **Bajar**, salvo la última.
  2. **Hechas**, debajo y sin botones: las historias Hechas, la que se completó más recientemente primero.
- **Subir o bajar** (`POST /proyectos/{id}/backlog/{numero}/subir` o `…/bajar`) mueve la historia un lugar dentro de su
  prioridad y vuelve (303) al backlog con el aviso "HU-3 subió un lugar." o "HU-3 bajó un lugar.".
- **La lista de candidatas** de la página de un sprint (HU-12) usa el mismo orden que el backlog (RN9).

## 4. Reglas de negocio

- RN1: **Orden del backlog.** Primero la prioridad (Alta, Media, Baja); dentro de una prioridad, el orden manual
  (`Orden`, de menor a mayor) y, si dos historias tienen el mismo, el número de historia (la más antigua primero). Las
  historias que ya existen tienen todas `orden` 0, así que hoy quedan ordenadas por número dentro de cada prioridad,
  como se ven ahora.
- RN2: **Subir o bajar mueve la historia un lugar** dentro de su prioridad, intercambiándola con la vecina inmediata. Las
  demás no cambian de lugar. **No se cruza a otra prioridad**: para pasar una historia de Media a Alta se cambia su
  prioridad (HU-06).
- RN3: **La primera de su prioridad no sube y la última no baja.** Una historia que es la única de su prioridad no
  puede moverse hacia ningún lado. Se informa con un error y no se cambia nada.
- RN4: **Las Hechas van aparte y no se reordenan.** No cuentan para el orden de las demás. Se muestran con la que se
  completó más recientemente primero (la fecha de HU-13) y, si la fecha es la misma, por número.
- RN5: **El orden se guarda.** Después de cada movimiento, las historias de esa prioridad que no están Hechas quedan con
  `orden` 1, 2, 3… consecutivos, todo en **una sola operación**: o se guardan todos o ninguno. Al volver a abrir la
  aplicación el orden es el mismo.
- RN6: **Una historia nueva va al final** de su prioridad: su `orden` es el más alto del proyecto más uno. Vale también
  después de haber reordenado a mano.
- RN7: **Cambiar la prioridad de una historia** (HU-06) **conserva su orden**: queda entre las historias de la nueva
  prioridad según ese orden (decisión 3). Se puede subir o bajar como cualquier otra.
- RN8: **Las historias asignadas a un sprint** (En sprint o En progreso) siguen en el backlog y se reordenan igual.
- RN9: **Las candidatas de un sprint** siguen el mismo orden que el backlog (decisión 5).

## 5. Restricciones

- **El orden y el movimiento son funciones puras** en `internal/domain/backlog`: `Ordenar` (RN1 y RN4: separa las
  pendientes de las Hechas y las ordena) y `Mover` (RN2, RN3 y RN5: devuelve el `orden` nuevo de las historias de esa
  prioridad). No usan base de datos ni HTTP, así se prueban con TDD. `Historia` suma el campo `Orden int`; `Nueva` le da el
  `orden` de una historia nueva (RN6).
- **Repositorios** (memoria y PostgreSQL): `Guardar` y `ListarPorProyecto` incluyen `orden`, y una operación nueva
  `ActualizarOrden` guarda el `orden` de varias historias del proyecto **en una transacción** (PostgreSQL) o de una vez
  (memoria), con el SQL parametrizado. **Sin migración nueva.**
- **Caso de uso** en `internal/service` (`Historias`): `Listar` devuelve el backlog ya separado y ordenado, y `Mover`
  busca la historia, aplica la regla y guarda el orden. Las candidatas de `Asignaciones.Ver` (HU-12) se ordenan con la misma
  función.
- **Pantalla:** los botones son formularios `POST` con redirección 303, como en el resto de la aplicación, sin JavaScript.
  Usa los componentes de TEC-07 (`exito`, `aviso`, `error`).
- **Sin roles:** cualquier usuario con sesión puede reordenar (TEC-05 solo da la identidad; quién puede qué queda para
  después).

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Backlog sin historias | La página dice que todavía no hay historias (como hoy) y no muestra la sección Hechas |
| Todas las historias están Hechas | La sección de pendientes dice que no queda nada pendiente y se muestra la de Hechas |
| Una historia sola en su prioridad | No tiene Subir ni Bajar |
| La primera de su prioridad pide subir (o la última, bajar) | Error 409 y no cambia nada (RN3) |
| Historias que ya existían, con `orden` 0 | Se ordenan por número; al primer movimiento de su prioridad quedan numeradas 1…n (RN5) |
| Dos historias con el mismo `orden` | Va primero la de número menor (RN1) |
| Una historia nueva después de reordenar | Va al final de su prioridad (RN6) |
| Cambiar la prioridad de una historia | Conserva su orden (RN7) |
| Una historia En sprint o En progreso | Se reordena igual (RN8) |
| Una historia Hecha | No se reordena (RN4) |
| Una historia de otro proyecto o que no existe | 404 |
| Dos personas mueven a la vez | Gana el último en guardar: el orden queda consistente porque se guarda todo junto (RN5) |

## 7. Condiciones de error

| Situación | Código | Mensaje al usuario |
|---|---|---|
| La primera de su prioridad pide subir | 409 | "HU-2 ya es la primera de su prioridad" |
| La última de su prioridad pide bajar | 409 | "HU-4 ya es la última de su prioridad" |
| La historia está Hecha | 409 | "una historia Hecha no se reordena" |
| La historia no existe en el proyecto | 404 | "no existe la historia HU-99" |
| El proyecto no existe | 404 | "No encontramos ese proyecto." (el de siempre) |
| Falla la base | 500 | "No se pudo guardar el orden. Probá de nuevo en unos minutos." |

## 8. Criterios de aceptación (issue #22)

- CA-09.1: Orden por prioridad y orden manual dentro de la misma prioridad.
- CA-09.2: El orden se persiste.
- CA-09.3: Las historias Hechas se muestran aparte.

## Decisiones a confirmar (antes de programar)

1. **Subir y bajar con botones**, un lugar por vez, y no arrastrar y soltar. Es lo que dice la tarea T4 del issue ("subir y
   bajar una historia"), no necesita JavaScript y se prueba con los tests de siempre. *Recomendado.*
2. **Orden por defecto: por número de historia** (la más antigua primero) dentro de cada prioridad, y una historia nueva va
   **al final** de su prioridad. Así el backlog de hoy se ve igual hasta que alguien reordena.
3. **Cambiar la prioridad (HU-06) conserva el orden de la historia.** Si una historia vieja pasa de Baja a Alta aparece entre
   las de Alta según su antigüedad (puede quedar arriba de otras más nuevas). La alternativa es mandarla al final de la nueva
   prioridad, pero eso obliga a cambiar la edición de HU-06 y no aporta mucho: después se sube o se baja. *Recomendado:
   conservarlo.*
4. **Las Hechas aparte**, en una sección debajo del backlog, con la completada más recientemente primero y sin botones.
5. **Las candidatas de la página del sprint (HU-12) usan el mismo orden.** Es el motivo de la historia ("decidir qué entra en
   el próximo sprint") y es un cambio chico en `Asignaciones.Ver`; hoy se ven por número. Si preferís dejar HU-12 como está,
   se saca la RN9 y su escenario. *Recomendado: incluirlo.*
6. **Sin migración y sin roles:** la columna `orden` ya existe y reordenar lo puede hacer cualquier usuario con sesión.

## Trazabilidad

Issue: #22 · Feature: `features/hu-09-priorizar-backlog.feature` · Pasos BDD: `features/hu09_steps_test.go`

Las columnas de código y tests se completan al implementar.
