# SPEC-17 · Abrir sesión de Planning Poker

**Historia:** HU-17 (#25) · **Responsable:** Piastrellini, Mariano · **Sprint:** 4 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa las tablas `sesiones_poker`, `rondas_poker` y `votos` que existen
> desde la migración 0001 (`docs/modelo-datos.md`), la escala de story points de HU-16 (`backlog.EscalaStoryPoints`),
> los integrantes de HU-03 y el usuario con sesión de TEC-05 (`Cuentas.Quien`). Tareas: #268 a #274.
>
> La sección **Modelo común de Planning Poker** (al final) es la base de HU-18, HU-19 y HU-22, que se hacen en paralelo
> (Planning del Sprint 4): dice qué tipos hay, qué repositorio usa cada historia y en qué archivo va cada cosa.

## 1. Objetivo

Que el Agile Enabler abra una sesión de Planning Poker para una historia del backlog, así el equipo la estima en
conjunto: la sesión queda abierta con su primera ronda y participan los integrantes activos del proyecto.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id (viene en la URL) | Sí | Tiene que existir |
| Historia | número HU-N (viene en la URL) | Sí | Tiene que existir en el proyecto y no estar Hecha (CA-17.1) |
| Quién abre | el integrante | Sí | Con login: el integrante vinculado a la cuenta (TEC-05). Sin login: se elige en el formulario (`quien`). Tiene que ser el Agile Enabler activo del proyecto (RN4) |

## 3. Salidas esperadas

- **Página de la historia** (`/proyectos/{id}/backlog/{numero}`): una sección **Planning Poker**.
  - Si la historia no tiene una sesión abierta y no está Hecha: el botón **Abrir Planning Poker** (sin login, con el
    desplegable **Quién abre**, que ofrece los integrantes activos).
  - Si ya tiene una abierta: el enlace **Ir a la sesión de Planning Poker**.
  - Si está Hecha: "La historia está Hecha: ya no se estima." y ningún botón.
- **Abrir** (`POST /proyectos/{id}/backlog/{numero}/poker`): crea la sesión **Abierta** con la **Ronda 1** y lleva a la
  página de la sesión con el mensaje **"Sesión de Planning Poker abierta."** (303).
- **Página de la sesión** (`GET /proyectos/{id}/backlog/{numero}/poker`): la historia (identificador, título y story
  points actuales), el estado de la sesión, la ronda actual ("Ronda 1") y la lista de **participantes**: los integrantes
  activos del proyecto, con su rol (CA-17.3). Desde HU-18 esta página tiene además el formulario para votar.

## 4. Reglas de negocio

- RN1: **Solo se abre para una historia que no está Hecha** (CA-17.1): Pendiente, En sprint o En progreso. Una historia
  que ya tiene story points se puede volver a estimar, como en HU-16.
- RN2: **Una única sesión abierta por historia** (CA-17.2). Abrir otra mientras hay una abierta es un error y la abierta
  sigue igual. Cuando HU-22 la cierra, se puede abrir una nueva.
- RN3: **La sesión se abre con la Ronda 1**, sin votos y sin revelar, en la misma operación: no existe una sesión
  abierta sin ronda.
- RN4: **La abre el Agile Enabler activo del proyecto** (el "Como" de la historia; el mismo rol que en HU-26 corrige
  todos los registros). Un Product Builder, un integrante dado de baja o alguien que no es integrante no la abre.
- RN5: **Participan los integrantes activos del proyecto** (CA-17.3), en cada momento: no se guarda una lista aparte.
  Uno que se da de baja deja de participar; uno que se registra con la sesión abierta participa desde ese momento.
- RN6: **Quién es quién** (decisión 1 de la Planning del Sprint 4): con login, el integrante vinculado a la cuenta con
  la que se entró (`Cuentas.Quien`, TEC-05); sin login (modo memoria y previews), el que se elige en el formulario. No
  hay tiempo real: la página se actualiza al recargar.

## 5. Restricciones

- **Sin migración:** las tablas existen desde la 0001, con el índice único parcial que garantiza RN2 en la base
  (`sesiones_poker_una_abierta_uq`) y la numeración única de las rondas.
- **Dominio puro** en `internal/domain/poker`: `PuedeAbrir` (RN1 y RN4) y `Participantes` (RN5). Sin base de datos ni
  HTTP.
- **Abrir es atómico** (RN3): en PostgreSQL, la sesión y la Ronda 1 se guardan en una transacción. Si dos pedidos abren
  a la vez, el índice único rechaza al segundo y el repositorio lo devuelve como `poker.ErrSesionYaAbierta`.
- **Pantalla** con los componentes de TEC-07, sin JavaScript.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Historia Pendiente, En sprint o En progreso | Se abre (RN1) |
| Historia que ya tiene story points | Se abre: se puede re-estimar (RN1) |
| Historia Hecha | No se abre (CA-17.1) y la página no ofrece el botón |
| Ya hay una sesión abierta | No se abre otra (CA-17.2): error con el enlace a la abierta |
| Dos pedidos de abrir a la vez | Se abre una sola (índice único) |
| Un integrante dado de baja | No aparece entre los participantes (CA-17.3) |
| Un integrante nuevo con la sesión abierta | Aparece entre los participantes al recargar (RN5) |
| El proyecto no tiene Agile Enabler activo | Nadie puede abrir: el mensaje dice que hay que registrarlo en Integrantes |
| Ver la sesión de una historia sin sesión abierta | 404 con el enlace a la historia |

## 7. Condiciones de error

| Situación | Error | Código | Mensaje al usuario |
|---|---|---|---|
| La historia está Hecha | `poker.ErrHistoriaHecha` | 409 | "La historia está Hecha: ya no se estima." |
| Ya hay una sesión abierta | `poker.ErrSesionYaAbierta` | 409 | "La historia ya tiene una sesión de Planning Poker abierta." y el enlace **Ir a la sesión** |
| Quien abre no es el Agile Enabler activo | `poker.ErrSoloAgileEnabler` | 403 | "Solo el Agile Enabler del proyecto abre una sesión de Planning Poker." |
| Sin login, no se eligió quién abre | `poker.ErrSoloAgileEnabler` | 422 | El mismo mensaje, debajo del desplegable |
| Con login, la cuenta no es integrante del proyecto | `member.ErrNoEsIntegrante` | 403 | "No sos integrante de este proyecto." |
| El proyecto o la historia no existen | — | 404 | Los de siempre |
| No hay sesión abierta (al verla) | `poker.ErrSesionNoEncontrada` | 404 | "Esta historia no tiene una sesión de Planning Poker abierta." |
| Falla la base | — | 500 | "No se pudo abrir la sesión. Probá de nuevo en unos minutos." |

## 8. Criterios de aceptación (issue #25)

- CA-17.1: Sólo para historias no Hechas.
- CA-17.2: Una única sesión abierta por historia.
- CA-17.3: Participan los integrantes activos del proyecto.

## Decisiones (confirmadas por Mariano el 26/10)

1. **Solo el Agile Enabler abre la sesión** (RN4), con `poker.PuedeConducir`, que reusan HU-19 (cerrar la ronda) y
   HU-22 (registrar el valor acordado).
2. **Los participantes no se guardan:** son los integrantes activos en cada momento (RN5). Así no hace falta migración.
3. **La Ronda 1 se crea al abrir**, en la misma transacción (RN3).
4. **Una página por sesión**, colgada de la historia (`/proyectos/{id}/backlog/{numero}/poker`), con el botón en la
   página de la historia.
5. **Sin migración:** alcanzan las tablas de la 0001 (la Planning preveía una; no hace falta).
6. **Dos PRs:** el primero, hoy, con esta spec, la feature y los tipos del modelo común, para que HU-18, HU-19 y HU-22
   arranquen sobre algo mergeado; el segundo, con el código de HU-17 (TDD).

## Modelo común de Planning Poker (HU-17, HU-18, HU-19 y HU-22)

### Tipos (`internal/domain/poker/poker.go`, en el primer PR)

| Tipo | Campos | Tabla |
|---|---|---|
| `Sesion` | `ID`, `HistoriaID`, `Estado` (`EstadoAbierta` o `EstadoCerrada`), `ValorAcordado *int` (nil hasta HU-22) | `sesiones_poker` |
| `Ronda` | `ID`, `SesionID`, `Numero` (desde 1), `Revelada` | `rondas_poker` |
| `Voto` | `ID`, `RondaID`, `IntegranteID`, `Valor` (de `backlog.EscalaStoryPoints`) | `votos` |

Errores comunes, en el mismo archivo: `ErrSesionNoEncontrada` y `ErrSesionCerrada`. Cada historia agrega los suyos en
su archivo (por ejemplo, HU-18 `ErrVotoFueraDeEscala`).

### Un archivo por historia

Para no pisarnos, cada historia agrega sus funciones y métodos en **sus propios archivos** del mismo paquete:

| Historia | Dominio (`internal/domain/poker`) | Caso de uso (`internal/service`) | Repositorios (`internal/store/memory` y `postgres`) |
|---|---|---|---|
| HU-17 | `abrir.go`: `PuedeAbrir`, `PuedeConducir`, `Participantes` | `poker.go`: `Poker.Abrir`, `Poker.Ver` | `poker.go`: el tipo `Poker`, `Abrir`, `SesionAbierta` |
| HU-18 | `votar.go` | `poker_votar.go` | `poker_votos.go`: `Votar` (revotar es un `UPDATE`) y `Votos` |
| HU-19 | `revelar.go` | `poker_revelar.go` | `poker_revelar.go`: `Revelar` |
| HU-22 | `acordar.go` | `poker_acordar.go` | `poker_acordar.go`: `Cerrar` y las sesiones cerradas (CA-22.3) |

### Repositorios: una interfaz por historia

Un solo tipo `Poker` en cada store implementa todo, pero cada caso de uso pide solo la interfaz de su historia, así
nadie cambia una interfaz ajena:

```go
// HU-17 (internal/service/poker.go)
type RepositorioSesionesPoker interface {
	// Abrir guarda una sesión Abierta de la historia con su Ronda 1, en una transacción. Si ya hay una abierta devuelve
	// poker.ErrSesionYaAbierta.
	Abrir(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error)
	// SesionAbierta devuelve la sesión abierta de la historia y su ronda actual (la de número más alto), o
	// poker.ErrSesionNoEncontrada.
	SesionAbierta(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error)
}
```

HU-18 agrega `RepositorioVotosPoker` (votar y listar los votos de una ronda), HU-19 `RepositorioRevelarPoker` y HU-22
`RepositorioAcordarPoker`, cada una en su archivo de `internal/service`.

### Página y quién es quién

- La página de la sesión es la de HU-17 (`web/templates/paginas/poker.html`, handler `poker` en
  `internal/server/poker.go`). HU-18 agrega el formulario de votar, HU-19 los votos revelados y HU-22 el valor
  acordado, cada uno en un bloque `{{define}}` propio para no tocar las mismas líneas.
- HU-17 deja en `internal/server/poker.go` la función **`quienParticipa`**, que devuelve el integrante que hace el
  pedido: con login, `Cuentas.Quien`; sin login, el `quien` del formulario. HU-18 la usa para votar.
- **Los votos no se muestran ni se devuelven antes de revelar** (CA-18.2 y CA-18.3): la vista de la sesión solo dice
  quién votó hasta que HU-19 revela la ronda.

## Trazabilidad

Issue: #25 · Feature: `features/hu-17-abrir-sesion-poker.feature` · Pasos BDD: `features/hu17_steps_test.go`

| Tarea | Qué | Dónde |
|---|---|---|
| T1 (#268) | Spec y escenarios | Esta spec y la feature |
| T2 (#269) | Dominio | `internal/domain/poker` |
| T3 (#270) | Repositorios | `internal/store/memory/poker.go` y `internal/store/postgres/poker.go` |
| T4 (#271) | Caso de uso | `internal/service/poker.go` |
| T5 (#272) | Pantalla | `internal/server/poker.go`, `web/templates/paginas/poker.html` y la sección de `historia.html` |
| T6 (#273) | App y BDD | `app/app.go` y `features/hu17_steps_test.go` |
| T7 (#274) | Documentación | Esta spec, la matriz de trazabilidad y el registro de IA |
