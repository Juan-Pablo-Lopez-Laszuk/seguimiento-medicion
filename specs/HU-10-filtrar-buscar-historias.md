# SPEC-10 · Filtrar y buscar historias

**Historia:** HU-10 (#23) · **Responsable:** Piastrellini, Mariano · **Sprint:** 3 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`. Filtra la página del backlog
> de HU-09 (ordenada por prioridad, con las Hechas aparte) y sigue las convenciones de los filtros de HU-29 (defectos):
> formulario `GET`, filtros combinables, un valor desconocido se ignora y un mensaje con enlace para quitar los filtros.

## 1. Objetivo

Que un integrante encuentre rápido una historia del backlog filtrando por estado y por prioridad y buscando por texto
en el título, sin perder el orden del backlog.

## 2. Entradas

Todas van en la URL de la página del backlog (`GET /proyectos/{id}/backlog?estado=…&prioridad=…&texto=…`), así una
búsqueda se puede recargar y compartir.

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Estado (`estado`) | uno de Pendiente, En sprint, En progreso, Hecho | No | Vacío o un valor que no existe = todos (RN4) |
| Prioridad (`prioridad`) | una de Alta, Media, Baja | No | Vacío o un valor que no existe = todas (RN4) |
| Texto (`texto`) | texto libre | No | Se quitan los espacios de los extremos; vacío = sin búsqueda |

## 3. Salidas esperadas

- Arriba de las historias, un **formulario** con los tres filtros (dos desplegables, "Todos" y "Todas" por defecto, y un
  campo de texto), el botón **Filtrar** y, si hay algún filtro puesto, el enlace **Quitar filtros**. El formulario muestra
  los valores que se están usando.
- Las historias que cumplen **todos** los filtros, **en el mismo orden y en las mismas secciones que sin filtrar**
  (Pendientes y en curso; Hechas). Una sección sin historias que cumplan no se muestra.
- Con algún filtro puesto, la cantidad: "3 de 10 historias".
- **Sin resultados:** "Ninguna historia coincide con los filtros." y el enlace **Quitar filtros** (CA-10.2). Es distinto
  del backlog vacío, que sigue diciendo "Todavía no hay historias en este proyecto.".

## 4. Reglas de negocio

- RN1: **Los filtros se combinan:** una historia aparece si cumple todos los que se eligieron (CA-10.1). Sin ningún
  filtro se ve el backlog completo.
- RN2: **La búsqueda es por el título** (CA-10.1): la historia aparece si su título contiene el texto, **sin distinguir
  mayúsculas ni tildes** ("sesion" encuentra "Iniciar sesión"; la ñ se distingue de la n). No busca en la descripción.
- RN3: **El orden y las secciones son los del backlog** (HU-09): filtrar solo saca historias, no las reordena. Las
  Hechas que cumplen siguen en su sección.
- RN4: **Un valor que no existe se ignora**, como en HU-29: `?estado=Cualquiera` muestra todos los estados. No es un
  error.
- RN5: **Con filtros no se reordena:** mientras haya algún filtro puesto no se muestran los botones Subir y Bajar, y se
  avisa "Para cambiar el orden, quitá los filtros." Mover con historias ocultas cambiaría el lugar respecto de
  historias que no se ven (decisión 4).
- RN6: **Los avisos de la página siguen igual** (historia creada, eliminada, movida): los filtros no los cambian.

## 5. Restricciones

- **El filtro es una función pura** en `internal/domain/backlog`: `Filtro` (estado, prioridad y texto), su constructor
  `NuevoFiltro` que ignora los valores desconocidos (RN4) y `Filtrar`, que devuelve las historias que cumplen sin cambiar
  el orden de la lista que recibe (RN3). Sin base de datos ni HTTP.
- **Se filtra en memoria después de leer el backlog**, igual en memoria y en PostgreSQL: el backlog de un proyecto tiene
  decenas de historias y así no hace falta otra consulta ni índices. **Sin migración.**
- **Caso de uso:** `Historias.Buscar` recibe el filtro y la vista suma el filtro usado y el total de historias del
  proyecto, para la cantidad "3 de 10". `Historias.Backlog` (HU-09) queda como `Buscar` sin filtros, así HU-09 no cambia.
- **Pantalla:** formulario `GET` sin JavaScript, con los componentes de TEC-07.
- La búsqueda sin tildes se hace con una tabla propia de equivalencias (á→a, é→e, í→i, ó→o, ú→u, ü→u), **sin
  dependencias nuevas**.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Sin filtros | El backlog completo, como en HU-09, con Subir y Bajar |
| Un filtro solo | Las historias que lo cumplen, en el orden del backlog |
| Varios filtros | Las que cumplen todos (RN1) |
| Estado Hecho | Solo la sección Hechas |
| Texto con mayúsculas, tildes o espacios en los extremos | Se ignoran (RN2) |
| Texto que solo tiene espacios | Es como no buscar |
| Ninguna historia cumple | "Ninguna historia coincide con los filtros." y Quitar filtros (CA-10.2) |
| Backlog vacío, con o sin filtros | "Todavía no hay historias en este proyecto." |
| Un estado o una prioridad que no existe en la URL | Se ignora (RN4) |
| Texto con caracteres especiales (`<script>`, `%`, `'`) | Se busca como texto común; la página lo muestra escapado |
| Filtros puestos | Sin Subir ni Bajar, con el aviso (RN5) |

## 7. Condiciones de error

| Situación | Código | Mensaje al usuario |
|---|---|---|
| El proyecto no existe | 404 | "No encontramos ese proyecto." (el de siempre) |
| Falla la base | 500 | "No se pudo cargar el backlog. Probá de nuevo en unos minutos." (el de siempre) |

Los filtros en sí no tienen errores: un valor desconocido se ignora (RN4).

## 8. Criterios de aceptación (issue #23)

- CA-10.1: Filtros por estado, prioridad y texto del título, combinables.
- CA-10.2: Sin resultados se muestra un mensaje claro.

## Decisiones (confirmadas por Mariano el 22/10)

1. **La rama sale de HU-09 (#421)**, que filtra su página: el PR apunta a la rama de HU-09 y pasa a `main` cuando se
   mergee #421.
2. **Un solo valor por filtro** (desplegables), como en HU-29.
3. **La búsqueda ignora mayúsculas y tildes** y busca solo en el título.
4. **Con filtros no se muestran Subir y Bajar** (RN5).
5. **Un valor desconocido se ignora** (RN4), como en HU-29.
6. **Filtrar en memoria, sin consultas nuevas ni migración.**

## Trazabilidad

Issue: #23 · Feature: `features/hu-10-filtrar-buscar-historias.feature` · Pasos BDD: `features/hu10_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Dominio | `internal/domain/backlog/filtrar.go` (`Filtro`, `NuevoFiltro`, `Filtro.Activo`, `Filtrar`) | `internal/domain/backlog/filtrar_test.go` |
| Caso de uso | `Historias.Buscar` y `VistaBacklog.Filtro` y `Total` (`internal/service/historias.go`) | `internal/service/filtrar_test.go` |
| Pantalla | `historias.listar` y `mostrarBacklog` (`internal/server/historias.go`) y el formulario de `web/templates/paginas/backlog.html` | `internal/server/filtrar_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-10.1 Filtros combinables | Sin filtros se ven todas las historias · Filtrar por estado o por prioridad · Buscar por texto en el título · Combinar los tres filtros · Un estado que no existe se ignora · Con filtros no se ofrecen Subir ni Bajar | `TestFiltrar`, `TestFiltrar_BusquedaSinMayusculasNiTildes`, `TestNuevoFiltro_IgnoraLosValoresQueNoExisten`, `TestFiltro_Activo`, `TestNuevoFiltro_GuardaLosValoresLimpios`, `TestFiltrar_NoCambiaLaListaQueRecibe`, `TestBuscar_FiltraLasDosSeccionesYCuentaElTotal`, `TestBuscar_SinFiltrosEsElBacklogCompleto`, `TestBuscar_ProyectoInexistente`, `TestBacklog_TieneElFormularioDeFiltros`, `TestBacklog_FiltrarPorPrioridad`, `TestBacklog_BuscarPorTexto`, `TestBacklog_FiltrarLasHechas`, `TestBacklog_FiltrosQueNoExistenSeIgnoran`, `TestBacklog_ElTextoBuscadoSeMuestraEscapado` |
| CA-10.2 Mensaje sin resultados | Ninguna historia coincide | `TestBacklog_FiltrosSinResultados`, `TestBacklog_VacioConFiltros` |

Los pasos BDD piden la página del backlog a la aplicación armada con repositorios en memoria y leen el HTML: prueban el
filtro de punta a punta, mensajes incluidos. El filtro se aplica después de leer el backlog, así que es el mismo con
PostgreSQL (cuyo `ListarPorProyecto` ya prueban los tests del repositorio).
