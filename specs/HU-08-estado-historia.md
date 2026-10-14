# SPEC-08 · Cambiar el estado de una historia

**Historia:** HU-08 (#21) · **Responsable:** Piastrellini, Mariano · **Sprint:** 2 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`, la columna `historias.estado` de
> `docs/modelo-datos.md` y las decisiones 5 y 6 de la Planning del Sprint 2.

## 1. Objetivo

Que el estado de una historia refleje su avance (Pendiente, En sprint, En progreso, Hecho) y que **una sola regla**
decida qué cambios de estado valen. Esa regla vive en `internal/domain/backlog`: HU-12 la usa para pasar una historia a
"En sprint" al asignarla (decisión 5), HU-13 para pasarla a "Hecho" desde el sprint activo (decisión 6) y HU-14 para
devolverla a Pendiente al cerrar el sprint. Esta historia escribe la regla, la guarda en la base y la muestra en la
página de la historia.

## 2. Entradas

**Cambiar el estado** (desde la página de la historia)

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Historia | número de la historia (HU-N, viene en la URL) | Sí | Tiene que existir en ese proyecto |
| Estado nuevo | uno de Pendiente, En sprint, En progreso, Hecho | Sí | Cualquier otro valor es un error |

Para el caso de uso (lo que usan HU-12, HU-13 y HU-14) las entradas son las mismas, más los criterios de la historia, que
el caso de uso busca solo.

## 3. Salidas esperadas

- **Si el cambio vale:** la historia queda con el estado nuevo, guardado. Desde la pantalla se vuelve a la página de la
  historia con el mensaje **"La historia pasó a En progreso"** (con el estado que corresponda).
- **Si no vale:** la historia **no cambia** y se informa por qué (RN6). En la pantalla se vuelve a mostrar la página de la
  historia con el aviso (409).
- **Página de la historia:** muestra el estado con un color por estado y, debajo, qué sigue: el botón **"Empezar a
  trabajarla"** si está En sprint; y una frase de qué falta si está Pendiente (se asigna a un sprint), En progreso (se
  completa desde el sprint activo, con todos sus criterios cumplidos) o Hecho (ya terminó).
- **Lista del backlog:** el estado de cada historia con un color distinto por estado.

## 4. Reglas de negocio

- RN1: Los estados son **Pendiente, En sprint, En progreso y Hecho** (CA-08.1). Toda historia nace Pendiente (HU-05 RN6).
- RN2: Las transiciones válidas son estas (CA-08.1) y **ninguna otra**:

  | Desde | Hacia | La hace |
  |---|---|---|
  | Pendiente | En sprint | HU-12, al asignar la historia a un sprint |
  | En sprint | En progreso | HU-08, desde la página de la historia |
  | En progreso | Hecho | HU-13, desde el sprint activo (RN4) |
  | En sprint | Pendiente | HU-14, al cerrar el sprint |
  | En progreso | Pendiente | HU-14, al cerrar el sprint |

  **No se puede saltear un paso**: de Pendiente no se pasa directo a En progreso ni a Hecho, y de En sprint no se pasa
  directo a Hecho (decisión 6 de la Planning). Tampoco se retrocede (por ejemplo de En progreso a En sprint).
- RN3: **Hecho es final**: una historia Hecha no pasa a ningún otro estado. Reabrirla no está en los criterios de esta
  historia; si hace falta, se agrega después con su propia historia. Las historias que no se terminan al cerrar el sprint
  vuelven a Pendiente (RN2); las Hechas se quedan Hechas.
- RN4: **Solo pasa a Hecho si tiene criterios y todos están cumplidos** (CA-08.2). Se usa `Progreso.Completo()` de HU-07: una
  historia **sin criterios no pasa a Hecho** (con `0/0` no está completa).
- RN5: Pedir el **mismo estado** en el que ya está la historia es un error ("ya está En progreso"), no un cambio que no
  hace nada. A diferencia de marcar un criterio (HU-07), una transición no es algo que se repita.
- RN6: Toda transición inválida devuelve un **error descriptivo** (CA-08.3): dice de qué estado a qué estado se quiso pasar
  y a cuáles se puede pasar desde ahí; si es por los criterios, cuántos están cumplidos.
- RN7: Si el cambio no vale, la historia **queda como estaba**: no se guarda nada.
- RN8: **Desde la página de la historia solo se ofrece En sprint → En progreso.** Pendiente → En sprint lo hace HU-12, que
  además guarda el sprint de la historia; En progreso → Hecho lo hace HU-13, que además exige que sea del sprint activo y
  guarda la fecha de completado (`completada_en`); y la vuelta a Pendiente la hace HU-14, que además guarda la foto del
  cierre. Si la pantalla ofreciera esos cambios, un estado quedaría sin su sprint o sin su fecha. Las tres usan la misma
  regla (RN2 a RN7), sin copiarla.
- RN9: Cambiar el estado **no toca** los criterios, las tareas, el sprint de la historia ni `completada_en`. Las tareas
  (HU-25) no condicionan el estado: los criterios de esta historia solo piden los criterios de aceptación.

## 5. Restricciones

- La regla es una **función pura** en `internal/domain/backlog` (`Historia.CambiarEstado`): recibe el estado nuevo y el
  progreso de los criterios (HU-07) y devuelve la historia con el estado nuevo o el error. No usa base de datos ni HTTP,
  así se prueba con TDD.
- El caso de uso en `internal/service` (`Estados.Cambiar`) busca el proyecto, la historia y los criterios, llama a la
  regla y guarda. **Cambios en los repositorios:** el de historias suma `ActualizarEstado` (memoria y PostgreSQL).
- **Sin migración nueva:** la columna `historias.estado` ya existe y su `CHECK` ya acepta los cuatro estados.
- El estado se cambia con un formulario `POST` y una redirección 303, como en el resto de la aplicación. La pantalla
  usa los componentes de TEC-07 (`exito`, `aviso`). Si en el pedido viene un estado que esta pantalla no ofrece (RN8),
  responde 400 y explica cuál se ofrece.
- Los estados nuevos (`EstadoEnSprint` y `EstadoEnProgreso`) se suman junto a `EstadoPendiente` y `EstadoHecho`, que ya
  existen.
- **Fuera de alcance**, porque tienen su propia historia: asignar a un sprint (HU-12), completar desde el sprint activo y
  la fecha de completado (HU-13), cerrar el sprint y devolver las historias a Pendiente (HU-14), y reabrir una historia
  Hecha.

## 6. Casos límite

Historia de ejemplo: HU-1 "Iniciar sesión con Google".

| Caso | Resultado |
|---|---|
| Pendiente → En sprint | Vale (lo usa HU-12) |
| En sprint → En progreso | Vale |
| En progreso → Hecho con 3 de 3 criterios cumplidos | Vale |
| En progreso → Hecho con 2 de 3 criterios cumplidos | Error: tiene 2 de 3 criterios cumplidos |
| En progreso → Hecho sin criterios | Error: no tiene criterios de aceptación |
| Pendiente → En progreso, Pendiente → Hecho | Error: no se puede saltear En sprint |
| En sprint → Hecho | Error: tiene que pasar por En progreso |
| En progreso → En sprint | Error: no se retrocede |
| En sprint → Pendiente, En progreso → Pendiente | Vale (lo usa HU-14 al cerrar el sprint) |
| Pendiente → Pendiente | Error: ya está Pendiente |
| Hecho → cualquier otro estado | Error: una historia Hecha no cambia de estado |
| Estado "Terminada" (no existe) | Error: el estado no existe |
| Un cambio inválido | La historia queda como estaba |
| Pedir desde la pantalla algo que no sea En progreso | 400: desde esa pantalla solo se pasa a En progreso |
| La historia no está En sprint y se pide En progreso | Error de transición, con el aviso en la página (409) |
| Proyecto o historia que no existe | Página "no encontrado" (404) |

## 7. Condiciones de error

| Situación | Error (Go) | Mensaje al usuario |
|---|---|---|
| Transición que no está en la tabla de RN2 | `ErrTransicionInvalida` | "no se puede pasar de Pendiente a Hecho: desde Pendiente solo se puede pasar a En sprint" |
| Desde Hecho | `ErrTransicionInvalida` | "no se puede pasar de Hecho a En progreso: una historia Hecha no cambia de estado" |
| Mismo estado | `ErrYaEstaEnEseEstado` | "la historia ya está En progreso" |
| Estado que no existe | `ErrEstadoDesconocido` | "el estado Terminada no existe: los estados son Pendiente, En sprint, En progreso y Hecho" |
| A Hecho sin criterios | `ErrCriteriosSinCumplir` | "no se puede pasar a Hecho: la historia no tiene criterios de aceptación" |
| A Hecho con criterios sin cumplir | `ErrCriteriosSinCumplir` | "no se puede pasar a Hecho: la historia tiene 2 de 3 criterios cumplidos y hacen falta todos" |
| La pantalla recibe un estado que no ofrece | — (lo controla el handler) | Página 400: "Desde esta pantalla solo se puede pasar a En progreso. A En sprint se pasa al asignar la historia a un sprint y a Hecho desde el sprint activo." |
| El proyecto no existe | `project.ErrNoEncontrado` | Página 404 "No encontramos ese proyecto" |
| La historia no existe en el proyecto | `backlog.ErrHistoriaNoEncontrada` | Página 404 "No encontramos esa historia" |

## 8. Criterios de aceptación

- CA-08.1: Transiciones válidas: Pendiente → En sprint → En progreso → Hecho (y vuelta a Pendiente al cerrar el sprint).
- CA-08.2: Solo pasa a Hecho si todos sus criterios de aceptación están cumplidos.
- CA-08.3: Una transición inválida devuelve un error descriptivo.

## Trazabilidad

Issue: #21 · Feature: `features/hu-08-estado-historia.feature` · Pasos BDD: `features/hu08_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Dominio | `internal/domain/backlog/estados.go` (`Estados`, `EstadoEnSprint`, `EstadoEnProgreso`, `Historia.CambiarEstado` y sus errores) | `internal/domain/backlog/estados_test.go` |
| Repositorio en memoria | `ActualizarEstado` en `internal/store/memory/historias.go` | `internal/store/memory/historias_test.go` |
| Repositorio PostgreSQL | `ActualizarEstado` en `internal/store/postgres/historias.go` | `internal/store/postgres/historias_test.go` |
| Caso de uso | `internal/service/estados.go` (`Estados.Cambiar`) | `internal/service/estados_test.go` |
| Pantalla | `internal/server/estados.go`, el bloque de estado de `web/templates/paginas/historia.html` y el color de cada estado en `historia.html` y `backlog.html` | `internal/server/estados_test.go` |
| Aplicación | `app/app.go` (elige memoria o PostgreSQL según `DATABASE_URL`) | `app/app_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-08.1 Transiciones válidas | Una historia En sprint pasa a En progreso · Una historia recorre todos los estados hasta Hecho · Qué transiciones valen | `TestEstados_EnElOrdenDelCamino`, `TestCambiarEstado_TransicionesValidas`, `TestCambiarEstado_Valido_GuardaElEstadoNuevo`, `TestCambiarEstado_ElCaminoCompletoHastaHecho`, `TestPaginaDeLaHistoria_EnSprint_OfreceEmpezarATrabajarla`, `TestCambiarEstado_EnSprintAEnProgreso_VuelveALaPaginaConElEstadoNuevo`, `TestListaDelBacklog_MuestraElEstadoConUnColorPorEstado` |
| CA-08.2 A Hecho solo con todos los criterios cumplidos | Para pasar a Hecho hacen falta todos los criterios cumplidos · Un cambio inválido deja la historia como estaba | `TestCambiarEstado_AHecho_ExigeTodosLosCriteriosCumplidos`, `TestCambiarEstado_AOtroEstado_NoMiraLosCriterios`, `TestCambiarEstado_AHechoConCriteriosSinCumplir_NoCambiaNada`, `TestCambiarEstado_AHechoSinCriterios_NoPasa`, `TestCambiarEstado_AHecho_UsaSoloLosCriteriosDeLaHistoria` |
| CA-08.3 Transición inválida con error descriptivo | Qué transiciones valen · Pasar a un estado en el que ya está · Pasar a un estado que no existe | `TestCambiarEstado_TransicionesInvalidas_DanUnErrorDescriptivo`, `TestCambiarEstado_ElMismoEstado_EsUnError`, `TestCambiarEstado_UnEstadoQueNoExiste_EsUnError`, `TestCambiarEstado_AHechoDesdeUnEstadoQueNoVale_InformaLaTransicion`, `TestCambiarEstado_TransicionInvalida_NoCambiaNada`, `TestCambiarEstado_Invalido_MuestraElAvisoYNoCambiaNada` |
| RN8 La pantalla solo ofrece En sprint → En progreso | — | `TestPaginaDeLaHistoria_DiceQueSigueEnCadaEstado`, `TestCambiarEstado_AUnEstadoQueLaPantallaNoOfrece_Responde400YNoCambiaNada` |
| Se guarda en la base | — | `TestHistorias_ActualizarEstado`, `TestHistorias_LosEstadosDelCodigoCoincidenConElCheckDeLaMigracion`, `TestNewHandlerConBase_ConURLElEstadoDeLaHistoriaSobreviveAUnReinicio` |
| Proyecto o historia inexistente (sección 7) | — | `TestCambiarEstado_ProyectoOHistoriaInexistente_DevuelveNoEncontrado`, `TestCambiarEstado_ProyectoOHistoriaInexistente_Responde404` |
