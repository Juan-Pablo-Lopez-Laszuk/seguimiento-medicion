# SPEC-29 · Consultar defectos

**Historia:** HU-29 (#7) · **Responsable:** Bravo, Carolina · **Sprint:** 3 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Tareas: #173 a #178.

## 1. Objetivo

Ver los defectos del proyecto filtrados, para priorizar cuáles arreglar.

## 2. Entradas

| Dato | De dónde sale | Validación |
|---|---|---|
| Proyecto | id (viene en la URL) | Tiene que existir |
| Severidad | `?filtro_severidad=` | Crítica, Alta, Media o Baja; vacío o cualquier otro valor = todas |
| Estado | `?filtro_estado=` | Abierto, En progreso, Resuelto, Cerrado o Reabierto; otro valor = todos |
| Sprint | `?filtro_sprint=` (id) | Sprint de detección, del proyecto; otro valor = todos |

## 3. Salidas esperadas

- En la **página de defectos**, un formulario de filtros (severidad, estado y sprint) arriba de la lista. La lista
  muestra solo los defectos que cumplen todos los filtros, el más nuevo primero (CA-29.1).
- Arriba de la lista, la **cantidad de defectos por severidad** de la lista filtrada: Crítica, Alta, Media y Baja,
  también las que dan 0 (CA-29.2).
- Si ningún defecto cumple los filtros: "Ningún defecto con esos filtros." y un enlace para quitarlos.

## 4. Reglas de negocio

- RN1: Los filtros se combinan: un defecto aparece si cumple todos los que se eligieron (CA-29.1).
- RN2: El filtro de sprint es por sprint de detección.
- RN3: Un valor que no existe o un sprint de otro proyecto se ignora: se muestran todos (no es un error).
- RN4: Los filtros de la lista no cambian el resumen de HU-34, que tiene su propio filtro de severidad (`?severidad=`).

## 5. Restricciones

- El filtro está en `internal/domain/defect` (`Filtro`, `Cumple`) y los repositorios lo aplican
  (`ListarFiltrados`: en memoria con `Cumple`, en PostgreSQL con parámetros en el `WHERE`, nunca armando SQL con
  texto del usuario). Caso de uso `Defectos.Consultar`. Solo lee. Sin migración nueva.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Sin filtros | Todos los defectos, como en HU-27 |
| Filtros que no coinciden con ninguno | Lista vacía con el aviso y todas las severidades en 0 |
| Estado "Hecho" en la URL | Se ignora: todos los estados |

## 7. Condiciones de error

| Condición | Error | Mensaje / respuesta |
|---|---|---|
| Proyecto inexistente | `project.ErrNoEncontrado` | Página 404 |

## 8. Criterios de aceptación

- CA-29.1: Filtros por severidad, estado y sprint.
- CA-29.2: Cantidad de defectos por severidad.

## Trazabilidad

Issue: #7 · Feature: `features/hu-29-consultar-defectos.feature` · Pasos BDD: `features/hu29_steps_test.go`

| Tarea | Capa | Código | Tests |
|---|---|---|---|
| T2 (#174) | Dominio y repositorios | `internal/domain/defect/filtro.go` (`Filtro`, `Cumple`, `Estados`); `ListarFiltrados` en `internal/store/memory/defectos.go` y `internal/store/postgres/defectos.go` | `filtro_test.go`, `defectos_test.go` de cada repositorio |
| T3 (#175) | Caso de uso | `internal/service/defectos.go` (`Consultar`, `VistaDefectos.PorSeveridad`) | `internal/service/defectos_test.go` (`TestConsultarDefectos`) |
| T4 (#176) | Pantalla | `internal/server/defectos.go` (`?filtro_severidad`, `?filtro_estado`, `?filtro_sprint`), `web/templates/paginas/defectos.html` | `TestPaginaDeDefectos_FiltrosYCantidadPorSeveridad` |
| T5 (#177) | Integración y BDD | `RepositorioDefectos` suma `ListarFiltrados` (la app ya lo arma) | `features/hu29_steps_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-29.1 Filtros | Filtrar por severidad, estado y sprint · Filtrar por estado y contar · Un estado que no existe se ignora | `TestFiltro_Cumple`, `TestDefectos_ListarFiltrados`, `TestConsultarDefectos` |
| CA-29.2 Cantidad por severidad | Filtrar por estado y contar · Ningún defecto cumple los filtros | `TestConsultarDefectos`, `TestPaginaDeDefectos_FiltrosYCantidadPorSeveridad` |
