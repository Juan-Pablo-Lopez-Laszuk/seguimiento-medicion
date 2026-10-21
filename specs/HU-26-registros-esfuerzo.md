# SPEC-26 · Consultar y corregir registros de esfuerzo

**Historia:** HU-26 (#4) · **Responsable:** Bravo, Carolina · **Sprint:** 3 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa los registros de HU-24, las tareas de HU-25 y la desviación de
> HU-33. Tareas: #152 a #158.

## 1. Objetivo

Que cada integrante pueda ver y corregir las horas que cargó, para que los datos de esfuerzo sean confiables.

## 2. Entradas

**Consulta** (`/proyectos/{id}/esfuerzo`):

| Dato | De dónde sale | Validación |
|---|---|---|
| Proyecto | id (viene en la URL) | Tiene que existir |
| Integrante | `?integrante=` (id) | Integrante del proyecto; otro valor = todos |
| Historia | `?historia=` (número HU-N) | Historia del proyecto; otro valor = todas |
| Sprint | `?sprint=` (número) | Sprint del proyecto; otro valor = todos |

**Corrección** (`/proyectos/{id}/esfuerzo/{registro}/corregir`):

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Registro | id (viene en la URL) | Sí | Tiene que existir y ser de una historia del proyecto |
| Quién corrige | id de un integrante | Sí | Integrante **activo** del proyecto (CA-26.2) |
| Tarea | id de una tarea | No | Si viene, tiene que ser de la historia del registro |
| Fecha, actividad y horas | como en HU-24 | Sí | Las mismas reglas de HU-24 |

## 3. Salidas esperadas

- **Página de registros de horas**, con botón "Registros de horas" en la vista del proyecto: un formulario de
  filtros (integrante, historia y sprint) y la lista de registros que cumplen todos los filtros, la fecha más
  reciente primero, con fecha, integrante, historia, tarea, actividad, horas y un enlace **Corregir**. Abajo, el total
  de horas de la lista.
- En la misma página, la tabla **horas estimadas vs reales por historia** (CA-26.3): una fila por historia de la lista,
  con horas estimadas (suma de sus tareas, M4), horas reales (todos sus registros, M5) y la desviación (M6), igual que
  la página de horas de HU-33.
- Si ningún registro cumple los filtros: "Ningún registro con esos filtros." y un enlace para quitarlos.
- **Corregir**: un formulario con los datos del registro y el campo "Quién corrige". Si se guarda, se vuelve a la
  lista con el mensaje **"Registro corregido."**.

## 4. Reglas de negocio

- RN1: Los filtros se combinan: un registro aparece si cumple todos los que se eligieron (CA-26.1).
- RN2: El filtro de sprint es por **fecha**: aparecen los registros con fecha entre el inicio y el fin del sprint.
  Así un registro queda en el sprint en que se trabajó aunque después la historia pase a otro sprint.
- RN3: Un valor que no existe o que es de otro proyecto se ignora: se muestran todos (no es un error).
- RN4: Cada integrante corrige solo sus registros; el Agile Enabler corrige todos (CA-26.2). Como la app no tiene
  usuarios con contraseña, quien corrige se elige en el formulario.
- RN5: Al corregir se pueden cambiar la tarea, la fecha, la actividad y las horas. El integrante y la historia del
  registro no cambian.
- RN6: La corrección valida lo mismo que HU-24 (horas, fecha, tarea y actividad). Para el máximo de 24 h en el día no
  se cuenta el mismo registro que se corrige.

## 5. Restricciones

- Dominio en `internal/domain/effort`: `Filtro` y `Cumple` (RN1 y RN2), `PuedeCorregir` (RN4) y `Corregir` (RN5 y
  RN6).
- Repositorios de registros: `ListarFiltrados`, `BuscarPorID` y `Actualizar`, en memoria y en PostgreSQL (con
  parámetros en el `WHERE`, nunca armando SQL con texto del usuario).
- Caso de uso nuevo `service.RegistrosEsfuerzo` (`Consultar`, `VerCorreccion` y `Corregir`). Las horas por historia
  usan la misma función que la página de horas de HU-33, así las dos páginas no pueden dar distinto.
- Sin migración nueva.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Sin filtros | Todos los registros del proyecto |
| Filtros que no coinciden con ninguno | Lista vacía con el aviso, total 0 h y sin filas en la tabla por historia |
| Corregir 20 h a 24 h en un día que tiene solo ese registro | Se guarda (no se cuenta dos veces) |
| El Agile Enabler corrige un registro de otro | Se guarda |

## 7. Condiciones de error

| Condición | Error | Mensaje / respuesta |
|---|---|---|
| Proyecto inexistente | `project.ErrNoEncontrado` | Página 404 |
| Registro inexistente o de otro proyecto | `effort.ErrRegistroNoEncontrado` | Página 404 |
| No se eligió quién corrige, o está dado de baja | `effort.ErrQuienInvalido` | "elegí quién corrige: un integrante activo del proyecto" |
| Un integrante corrige el registro de otro | `effort.ErrNoPuedeCorregir` | "cada integrante corrige solo sus registros (el Agile Enabler, todos)" |
| Horas, fecha, tarea o actividad inválidas | los de HU-24 | Los mensajes de HU-24, debajo de cada campo |

## 8. Criterios de aceptación

- CA-26.1: Se puede filtrar por integrante, historia y sprint.
- CA-26.2: Cada uno edita solo sus registros (el Agile Enabler, todos).
- CA-26.3: Muestra horas estimadas vs reales por historia.

## Trazabilidad

Issue: #4 · Feature: `features/hu-26-registros-esfuerzo.feature` · Pasos BDD: `features/hu26_steps_test.go`

| Tarea | Capa | Código | Tests |
|---|---|---|---|
| T2 (#153) | Dominio | `internal/domain/effort/correccion.go` (`Filtro`, `Cumple`, `PuedeCorregir`, `Corregir`) | `correccion_test.go` |
| T3 (#154) | Repositorios | `ListarFiltrados`, `BuscarPorID` y `Actualizar` en `internal/store/memory/registros.go` y `internal/store/postgres/registros.go` | `registros_hu26_test.go` de cada repositorio |
| T4 (#155) | Caso de uso | `internal/service/registros_esfuerzo.go` (`RegistrosEsfuerzo`: `Consultar`, `VerCorreccion`, `Corregir`) | `internal/service/registros_esfuerzo_test.go` |
| T5 (#156) | Pantalla | `internal/server/registros_esfuerzo.go`, `web/templates/paginas/registros_esfuerzo.html`, `corregir_registro.html`, botón en `proyecto.html` | `internal/server/registros_esfuerzo_test.go` |
| T6 (#157) | Integración y BDD | `app/app.go` arma `RegistrosEsfuerzo`; `RepositorioRegistros` suma los métodos de HU-26 | `app/registros_esfuerzo_test.go`, `features/hu26_steps_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-26.1 Filtros | Filtrar por integrante y sprint · Ver horas estimadas y reales por historia · Un integrante que no existe se ignora | `TestFiltro_Cumple`, `TestRegistros_ListarFiltrados`, `TestConsultarRegistros_Filtros`, `TestPaginaDeRegistros_FiltrosListaYHorasPorHistoria` |
| CA-26.2 Corregir solo lo propio (el Agile Enabler, todo) | Corregir un registro propio · El Agile Enabler corrige el registro de otro · Un integrante no corrige el registro de otro · Las horas corregidas siguen las reglas de HU-24 | `TestPuedeCorregir`, `TestCorregir_*`, `TestCorregirRegistro`, `TestCorregirRegistro_DeOtro_MuestraElError` |
| CA-26.3 Estimadas vs reales por historia | Ver horas estimadas y reales por historia · Corregir un registro propio | `TestConsultarRegistros_FilasYHorasPorHistoria`, `TestPaginaDeRegistros_FiltrosListaYHorasPorHistoria` |
