# SPEC-29 · Consultar defectos

**Historia:** HU-29 (#7) · **Responsable:** Bravo, Carolina · **Sprint:** 3 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Tareas: #173 a #178.

## 1. Objetivo

Ver los defectos del proyecto filtrados, para priorizar cuáles arreglar.

## 2. Entradas

| Dato | De dónde sale | Validación |
|---|---|---|
| Proyecto | id (viene en la URL) | Tiene que existir |
| Severidad | `?severidad=` | Crítica, Alta, Media o Baja; vacío o cualquier otro valor = todas |
| Estado | `?estado=` | Abierto, En progreso, Resuelto, Cerrado o Reabierto; otro valor = todos |
| Sprint | `?sprint=` (id) | Sprint de detección, del proyecto; otro valor = todos |

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
- RN4: La severidad elegida también filtra el resumen de HU-34 (es el mismo parámetro de la página).

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
