# SPEC-22 · Registrar la estimación acordada

**Historia:** HU-22 (#30) · **Responsable:** López Laszuk, Juan Pablo · **Sprint:** 4 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Cierra la sesión de Planning Poker que abre HU-17, sobre el
> **Modelo común de Planning Poker** de `specs/HU-17-abrir-sesion-poker.md`. Guarda la estimación con la misma regla que
> HU-16 (`backlog.Historia.Estimar`) y anota el cambio en el historial de HU-16b. Usa `ValorAcordado` de `poker.Sesion`,
> `Revelada` de `poker.Ronda` (HU-19) y `poker.PuedeConducir` y `quienParticipa` (HU-17). Tareas: #303 a #309.

## 1. Objetivo

Que el Agile Enabler, después de ver los votos, registre el valor que acordó el equipo. Ese valor queda como estimación
de la historia y la sesión se cierra. Las sesiones cerradas se pueden consultar en la página de Planning Poker de la
historia.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id (viene en la URL) | Sí | Tiene que existir |
| Historia | número HU-N (viene en la URL) | Sí | Tiene que existir en el proyecto, no estar Hecha y tener una sesión abierta |
| Valor acordado | entero (desplegable `valor` con la escala) | Sí | Tiene que ser de la escala `0, 1, 2, 3, 5, 8, 13, 21` (CA-22.1) |
| Quién registra | el integrante | Sí | Con login: el integrante vinculado a la cuenta. Sin login: se elige en el formulario (`quien`). Tiene que ser el Agile Enabler activo (RN3) |

## 3. Salidas esperadas

- **Página de la sesión** (`/proyectos/{id}/backlog/{numero}/poker`, de HU-17), con la sesión abierta: el bloque
  **Registrar la estimación** (`{{define "pokerAcordar"}}`, que llama `pokerRonda`).
  - Si la ronda actual está revelada: el desplegable **Valor acordado** con la escala y el botón **Registrar
    estimación**. Sin login, también el desplegable **Quién registra**.
  - Si todavía no está revelada: "Para registrar la estimación, primero revelá la ronda."
- **Registrar** (`POST /proyectos/{id}/backlog/{numero}/poker/acordar`), en una sola operación:
  - deja la sesión **Cerrada** con su valor acordado (CA-22.3);
  - pone el valor como story points de la historia (CA-22.2);
  - vuelve a la página de la sesión (`?acordada=5`) con el mensaje **"Estimación acordada: 5 SP. La sesión de Planning
    Poker quedó cerrada."** (303). El mensaje solo sale si el valor de la URL es de la escala.
- **Sesiones cerradas** (CA-22.3): en la misma página, debajo de la sesión, la tabla **Sesiones cerradas** con la más
  reciente primero (`{{define "pokerCerradas"}}`). Columnas: Sesión ("Sesión 1", "Sesión 2", en orden de apertura),
  Rondas y Valor acordado. Sin sesiones cerradas: "Todavía no se cerró ninguna sesión."

## 4. Reglas de negocio

- RN1: **El valor es de la escala** (CA-22.1). Lo valida `backlog.Historia.Estimar`, la misma regla que HU-16. El valor
  acordado no tiene que coincidir con ningún voto: es lo que decide el equipo después de discutir.
- RN2: **Los story points de la historia pasan a ser el valor acordado** (CA-22.2).
  - Si la historia ya tenía otra estimación, el cambio se anota en el historial de HU-16b.
  - Si el valor es el mismo que ya tenía, la estimación y el historial no cambian, pero la sesión se cierra igual.
- RN3: **Lo registra el Agile Enabler activo del proyecto** (`poker.PuedeConducir`, la misma regla que abre la sesión en
  HU-17).
- RN4: **Solo con la ronda actual revelada** (HU-19): el valor se acuerda después de ver los votos. Que la ronda tenga
  votos no se exige.
- RN5: **La sesión queda cerrada y consultable** (CA-22.3).
  - No se acuerda dos veces en la misma sesión (`poker.ErrSesionCerrada`).
  - Desde ese momento se puede abrir una nueva sesión para la historia (RN2 de HU-17).
- RN6: **Todo o nada**: el cierre de la sesión, los story points y la entrada del historial se guardan juntos. Si falla
  uno, la sesión sigue abierta y la historia como estaba.
- RN7: **Una historia Hecha no se estima** (RN6 de HU-16). Si la historia pasó a Hecha con la sesión abierta, registrar
  da error y la sesión queda abierta.

## 5. Restricciones

- **Sin migración:** `sesiones_poker.valor_acordado` y su `CHECK` de la escala existen desde la 0001. La tabla no guarda
  fechas, así que las sesiones cerradas se numeran y se ordenan por cómo se abrieron (`id`).
- **Dominio puro** en `internal/domain/poker/acordar.go`: `Acordar(historia, sesion, ronda, quien, valor)` devuelve la
  historia estimada. Revisa, en este orden, que la sesión esté abierta (RN5), quién registra (RN3) y la ronda revelada
  (RN4); el valor y la historia Hecha los decide `backlog.Historia.Estimar` (RN1 y RN7). También define
  `SesionCerrada` (número, rondas y valor).
- **Caso de uso** `Poker.Acordar` en `internal/service/poker_acordar.go`, con su propia interfaz de repositorio, que se
  conecta con `Poker.ConAcordar`. `Poker.Ver` suma las sesiones cerradas (`VistaSesion.Cerradas`).

```go
// HU-22 (internal/service/poker_acordar.go)
type RepositorioAcordarPoker interface {
	// SesionAbierta devuelve la sesión abierta de la historia y su ronda actual, o poker.ErrSesionNoEncontrada.
	SesionAbierta(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error)
	// Cerrar deja la sesión Cerrada con el valor acordado y guarda los story points de la historia (con su entrada del
	// historial de HU-16b si cambian), en una transacción. Si la sesión ya no está abierta devuelve
	// poker.ErrSesionCerrada.
	Cerrar(ctx context.Context, sesionID, historiaID int64, valor int) error
	// SesionesCerradas devuelve las sesiones cerradas de la historia, la más reciente primero, con su cantidad de rondas.
	SesionesCerradas(ctx context.Context, historiaID int64) ([]poker.SesionCerrada, error)
}
```

- **PostgreSQL:** `Cerrar` hace `UPDATE sesiones_poker ... WHERE id = $1 AND estado = 'Abierta'` en la misma transacción
  que los story points y el historial. Si no actualiza ninguna fila, devuelve `poker.ErrSesionCerrada`: si dos pedidos
  llegan a la vez, el segundo espera al primero y después no encuentra la fila abierta. Los story points y el historial
  se escriben con `guardarStoryPoints`, la misma función que usa `Historias.ActualizarStoryPoints`, para no duplicar la
  regla de HU-16b.
- **Memoria:** `Poker.ConHistorias` conecta el repositorio de historias. `Cerrar` cierra la sesión con el candado (así de
  dos pedidos a la vez se registra uno) y después guarda los story points; si eso falla, la vuelve a abrir.
- **Pantalla** con los componentes de TEC-07, sin JavaScript.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Historia sin estimar | Queda con el valor acordado; no se anota en el historial (RN2 de HU-16b) |
| Historia con 3 SP, se acuerda 8 | Queda con 8 SP y el historial anota 3 → 8 (RN2) |
| Se acuerda el mismo valor que ya tenía | La sesión se cierra; la estimación y el historial no cambian (RN2) |
| Se acuerda 0 | Vale: 0 es una estimación (RN4 de HU-16) |
| Valor que no es de la escala (por ejemplo 4) | Error; la sesión sigue abierta y la historia como estaba (CA-22.1) |
| La ronda actual no está revelada | Error; la página no ofrece el formulario (RN4) |
| Registrar dos veces | El segundo da "Esta historia no tiene una sesión de Planning Poker abierta." (RN5) |
| Dos pedidos a la vez | Se registra uno; el otro recibe `ErrSesionCerrada` (RN5) |
| Después de cerrar, se abre otra sesión | Se abre (RN2 de HU-17); la cerrada sigue en la tabla |
| Dos sesiones cerradas | Se ven las dos, la más reciente primero, con su valor y sus rondas (CA-22.3) |
| La historia pasó a Hecha con la sesión abierta | Error; la sesión sigue abierta (RN7) |
| `?acordada=` con algo que no es de la escala | La página no muestra ningún mensaje |

## 7. Condiciones de error

| Situación | Error | Código | Mensaje al usuario |
|---|---|---|---|
| El valor no es de la escala | `domain.ErroresValidacion` (campo `story_points`) | 422 | "los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21", debajo del desplegable (el mismo de HU-16) |
| La ronda actual no está revelada | `poker.ErrAcordarSinRevelar` | 409 | "Para registrar la estimación, primero revelá la ronda." |
| Quien registra no es el Agile Enabler activo | `poker.ErrSoloAgileEnablerAcuerda` | 403 | "Solo el Agile Enabler del proyecto registra la estimación acordada." |
| Sin login, no se eligió quién registra | `poker.ErrSoloAgileEnablerAcuerda` | 422 | El mismo mensaje, debajo del desplegable |
| Con login, la cuenta no es integrante del proyecto | `member.ErrNoEsIntegrante` | 403 | "No sos integrante de este proyecto." |
| La historia no tiene una sesión abierta | `poker.ErrSesionNoEncontrada` | 404 | "Esta historia no tiene una sesión de Planning Poker abierta." |
| La sesión se cerró mientras tanto | `poker.ErrSesionCerrada` | 409 | "La sesión de Planning Poker ya está cerrada." |
| La historia está Hecha | `backlog.ErrEstimacionDeHistoriaHecha` | 409 | "La historia está Hecha: no se puede cambiar su estimación." |
| El proyecto o la historia no existen | — | 404 | Los de siempre |
| Falla la base | — | 500 | "No se pudo registrar la estimación. Probá de nuevo en unos minutos." |

## 8. Criterios de aceptación (issue #30)

- CA-22.1: El valor pertenece a la escala.
- CA-22.2: Se actualizan los SP de la historia.
- CA-22.3: La sesión queda cerrada y consultable.

## Decisiones (confirmadas por Juan Pablo el 26/10)

1. **Hace falta la ronda revelada** (RN4). Hasta que llegue HU-19, que revela la ronda, la sesión no se puede cerrar
   desde la app: los tests y los escenarios la revelan directo en el repositorio.
2. **El valor es libre dentro de la escala** (RN1): no tiene que ser uno de los votos.
3. **"Consultable" es la tabla de sesiones cerradas** (CA-22.3), con su valor y sus rondas, sin una página aparte con los
   votos de cada ronda. Va en la página de Planning Poker de la historia y no en la de la historia, porque HU-17 dejó
   todo Planning Poker en esa página y en la de la historia solo un enlace.
4. **El mismo valor cierra la sesión sin tocar el historial** (RN2).
5. **Sin migración:** sin fecha de cierre, porque la tabla no la tiene.

## Trazabilidad

Issue: #30 · Feature: `features/hu-22-registrar-estimacion.feature` · Pasos BDD: `features/hu22_steps_test.go`

| Tarea | Código | Tests |
|---|---|---|
| T1 (#303) | Esta spec y la feature | — |
| T2 (#304) | `internal/domain/poker/acordar.go` (`Acordar`, `SesionCerrada`, `ErrAcordarSinRevelar`, `ErrSoloAgileEnablerAcuerda`) | `internal/domain/poker/acordar_test.go` |
| T3 (#305) | `poker_acordar.go` en `internal/store/memory` (`ConHistorias`, `Cerrar`, `SesionesCerradas`) y en `internal/store/postgres` (`Cerrar`, `SesionesCerradas`); `guardarStoryPoints` en `internal/store/postgres/historias.go` | `poker_acordar_test.go` en los dos stores |
| T4 (#306) | `internal/service/poker_acordar.go` (`RepositorioAcordarPoker`, `ConAcordar`, `Acordar`) y `VistaSesion.Cerradas` en `poker.go` | `internal/service/poker_acordar_test.go` |
| T5 (#307) | `internal/server/poker_acordar.go` (`acordar`, `mensajeAcordada`), la ruta en `router.go` y los bloques `pokerAcordar` y `pokerCerradas` de `poker.html` | `internal/server/poker_acordar_test.go` |
| T6 (#308) | `app/app.go` y `features/hu22_steps_test.go` | `app/poker_acordar_test.go` y los escenarios |
| T7 (#309) | Esta spec, la matriz de trazabilidad y el registro de IA | — |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-22.1 El valor pertenece a la escala | Un valor fuera de la escala no se registra · Cero es una estimación | `TestAcordar_ElValorTieneQueSerDeLaEscala`, `TestAcordarPoker_ValorFueraDeLaEscala`, `TestPoker_AcordarSinLogin_Errores`, `TestPoker_AcordarSinLogin_LosErroresVanEnElCampo` |
| CA-22.2 Se actualizan los SP de la historia | El Agile Enabler registra la estimación acordada · Re-estimar con Planning Poker queda en el historial · Acordar el mismo valor cierra la sesión sin anotar nada | `TestAcordar_ConLaRondaRevelada_EstimaLaHistoria`, `TestAcordar_HistoriaYaEstimada_ReemplazaLaEstimacion`, `TestPoker_Cerrar_CierraLaSesionYEstimaLaHistoria`, `TestPoker_Cerrar_AnotaLaReestimacion`, `TestAcordarPoker_ElAgileEnablerRegistraLaEstimacion`, `TestPoker_AcordarSinLogin`, `TestNewHandlerConBase_ConURLRegistrarEstimacion` |
| CA-22.3 La sesión queda cerrada y consultable | El Agile Enabler registra la estimación acordada · No se registra dos veces · Las sesiones cerradas se consultan desde la historia | `TestPoker_SesionesCerradas`, `TestPoker_Cerrar_SesionYaCerrada`, `TestPoker_CerrarDosVecesALaVez`, `TestVerPoker_TraeLasSesionesCerradas`, `TestAcordarPoker_SinSesionAbierta`, `TestPoker_SinSesionesCerradas` |
| RN3 Quién registra | Un Product Builder no registra la estimación | `TestAcordar_SoloElAgileEnabler`, `TestAcordarPoker_ErroresDeLasReglas`, `TestPoker_ConLogin_RegistraElIntegranteDeLaCuenta`, `TestPoker_ConLogin_Acordar_Errores` |
| RN4 Con la ronda revelada | No se registra con la ronda sin revelar | `TestAcordar_ConLaRondaSinRevelar_NoSeAcuerda`, `TestPoker_SinRevelar_NoOfreceRegistrar`, `TestPoker_ConLaRondaRevelada_OfreceRegistrar`, `TestNewHandlerConBase_SinURLRegistrarEstimacion` |
| RN6 Todo o nada | — | `TestPoker_Cerrar_LoQueFallaNoCierraLaSesion` (memoria y PostgreSQL), `TestPoker_Acordar_FallaLaBase` |
| RN7 Historia Hecha | Una historia que pasó a Hecha no se estima | `TestAcordar_HistoriaHecha_NoSeEstima`, `TestAcordarPoker_HistoriaHecha` |
