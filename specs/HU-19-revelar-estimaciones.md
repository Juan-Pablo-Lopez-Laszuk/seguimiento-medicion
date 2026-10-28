# SPEC-19 · Revelar estimaciones

**Historia:** HU-19 (#27) · **Responsable:** Bravo, Carolina · **Sprint:** 4 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa las tablas `sesiones_poker`, `rondas_poker` y `votos` de
> `docs/modelo-datos.md` (migración 0001) y se apoya en HU-17 (abrir la sesión) y HU-18 (votar de forma oculta), de
> Mariano. Tareas: #283 a #289.
>
> Sigue el **modelo común de Planning Poker** de la spec de HU-17 (tipos `Sesion`, `Ronda` y `Voto` en
> `internal/domain/poker/poker.go`, un archivo por historia y una interfaz de repositorio por historia).

## 1. Objetivo

Ver todas las estimaciones de una ronda al terminar la votación, para que el equipo las discuta.

## 2. Entradas

| Dato | De dónde sale | Validación |
|---|---|---|
| Proyecto, historia y sesión | La URL de la sesión de Planning Poker (HU-17) | Tienen que existir; la sesión, Abierta |
| Ronda | La ronda actual de la sesión (`rondas_poker`, la de número más alto) | Tiene que existir |
| Votos | `votos` de la ronda (HU-18) | — |
| Integrantes | Los integrantes activos del proyecto (HU-03) | Son los que votan |
| Quién cierra la ronda | El integrante de la sesión iniciada (TEC-05) o el elegido en la pantalla sin login (decisión 1 de la Planning del Sprint 4) | Tiene que ser el Agile Enabler |

## 3. Salidas esperadas

- **Antes de revelar:** la pantalla de la sesión muestra quién votó y quién falta, sin los valores (regla de HU-18). El
  Agile Enabler ve el botón **"Cerrar la ronda y revelar"**.
- **Revelada:** la tabla **Votos de la ronda N**, con cada integrante y su voto (CA-19.2), ordenada por nombre. Los que
  no votaron aparecen con "no votó".
- Al revelarse sola (todos votaron) o al cerrarla, se vuelve a la sesión con el mensaje **"Ronda revelada."**.
- Las otras pestañas con la sesión abierta se enteran solas: el fragmento de quién votó de HU-18, que HTMX pide cada 5
  segundos, revisa la ronda (RN1) y, si está revelada, responde `HX-Refresh: true` para que la página se recargue y
  muestre la tabla de votos.

## 4. Reglas de negocio

- RN1: Una ronda se revela **sola** cuando votaron todos los integrantes activos del proyecto (CA-19.1). Se comprueba
  después de cada voto (HU-18) y al abrir la pantalla.
- RN2: El **Agile Enabler** puede cerrar la ronda antes: se revela con los votos que haya (CA-19.1).
- RN3: Una ronda **sin votos** no se revela, ni sola ni cerrándola (CA-19.3).
- RN4: Solo el Agile Enabler cierra la ronda.
- RN5: Una ronda revelada no se vuelve a revelar ni recibe votos nuevos. Para votar de nuevo se abre otra ronda (HU-21).
- RN6: Los valores de los votos solo se muestran en una ronda revelada (CA-19.2 y regla de HU-18).
- RN7: Solo cuentan y se muestran los votos de los integrantes activos: el de alguien que se dio de baja no cuenta para
  revelar ni para cerrar, y no aparece en la tabla (RN2 de HU-18).

## 5. Restricciones

Según el modelo común de HU-17, todo lo de HU-19 va en sus propios archivos:

| Capa | Archivo | Qué |
|---|---|---|
| Dominio | `internal/domain/poker/revelar.go` | `TodosVotaron`, `RevelarSiCorresponde`, `Cerrar` y `VotosRevelados`, funciones puras con TDD |
| Repositorios | `internal/store/memory/poker_revelar.go` y `internal/store/postgres/poker_revelar.go` | `Revelar`: marca la ronda como revelada (`rondas_poker.revelada`); `Votos`: los votos de la ronda con su valor (HU-18 solo dice quiénes votaron). Sin migración nueva: las columnas ya existen |
| Caso de uso | `internal/service/poker_revelar.go` | La interfaz `RepositorioRevelarPoker`; revelar sola después de cada voto y al ver la sesión, y cerrar la ronda |
| Pantalla | `internal/server/poker.go` y un bloque `{{define}}` propio en `web/templates/paginas/poker.html` | El botón "Cerrar la ronda y revelar" y la tabla de votos; quién cierra sale de `quienParticipa` (HU-17) |

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Votaron todos menos uno | No se revela sola; el Agile Enabler puede cerrarla |
| Un integrante dado de baja no votó | No cuenta: se revela si votaron todos los activos |
| El Agile Enabler cierra con un solo voto | Se revela con ese voto |
| Alguien vota y después se da de baja | Su voto no cuenta ni se muestra, como dice HU-18 (RN2 de HU-18); si era el último que faltaba, la ronda se revela con los votos de los activos |

## 7. Condiciones de error

| Condición | Error | Mensaje |
|---|---|---|
| Ronda sin votos | `poker.ErrRondaSinVotos` | "la ronda no tiene votos: no se puede revelar" |
| Cierra alguien que no es el Agile Enabler | `poker.ErrSoloAgileEnabler` | "solo el Agile Enabler del proyecto abre una sesión de Planning Poker" (el de `poker.PuedeConducir`, de HU-17) |
| La ronda ya estaba revelada | `poker.ErrRondaYaRevelada` | "la ronda ya está revelada" |
| Se piden los votos de una ronda sin revelar | `poker.ErrRondaSinRevelar` | "los votos se ven cuando la ronda se revela" |

## 8. Criterios de aceptación

- CA-19.1: Se revela cuando todos votaron o cuando el Agile Enabler cierra la ronda.
- CA-19.2: Se muestra cada voto con su autor.
- CA-19.3: Una ronda sin votos no se puede revelar.

## Trazabilidad

Issue: #27 · Feature: `features/hu-19-revelar-estimaciones.feature` · Pasos BDD: `features/hu19_steps_test.go`

| Tarea | Capa | Código | Tests |
|---|---|---|---|
| T1 (#283) | Spec y BDD | Esta spec y `features/hu-19-revelar-estimaciones.feature` | — |
| T2 (#284) | Dominio | `internal/domain/poker/revelar.go` (`TodosVotaron`, `RevelarSiCorresponde`, `Cerrar`, `VotosRevelados`) | `internal/domain/poker/revelar_test.go` |
| T3 (#285) | Repositorios | `internal/store/memory/poker_revelar.go` y `internal/store/postgres/poker_revelar.go` (`Revelar`) | `TestPoker_RevelarGuardaLaRondaRevelada`, `TestPoker_RevelarUnaRondaQueNoExiste` (en los dos) |
| T4 (#286) | Caso de uso | `internal/service/poker_revelar.go` (`RepositorioRevelarPoker`, `RevelarPoker.Revisar` y `RevelarPoker.Cerrar`) | `internal/service/poker_revelar_test.go` |
| T5 (#287) | Pantalla | `internal/server/poker_revelar.go` (`cerrar`, `conVotosRevelados`), bloque `pokerRevelar` en `poker.html` | `internal/server/poker_revelar_test.go` |
| T6 (#288) | Integración y BDD | `app/app.go` arma `RevelarPoker` con el mismo repositorio de HU-17 y HU-18 | `app/poker_revelar_test.go`, `features/hu19_steps_test.go` (votan con `Poker.Votar` de HU-18) |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-19.1 Se revela cuando votan todos o cierra el Agile Enabler | La ronda se revela cuando votaron todos · El Agile Enabler cierra la ronda antes de que voten todos · Falta un voto y la ronda sigue oculta · Un Product Builder no puede cerrar la ronda | `TestRevelarSiCorresponde`, `TestCerrar`, `TestRevelarPoker_SeRevelaSolaCuandoVotaronTodos`, `TestRevelarPoker_ElAgileEnablerCierraLaRonda`, `TestRevelar_VotaronTodos_SeRevelaSola`, `TestRevelar_ElAgileEnablerCierraLaRonda` |
| CA-19.2 Cada voto con su autor | La ronda se revela cuando votaron todos · El Agile Enabler cierra la ronda antes de que voten todos | `TestVotosRevelados`, `TestRevelarPoker_ElAgileEnablerCierraLaRonda` |
| CA-19.3 Sin votos no se revela | Una ronda sin votos no se puede revelar | `TestCerrar`, `TestRevelarPoker_CerrarConError`, `TestRevelar_CerrarConError` |

**Revisión de Mariano después del merge (#435):**

| Arreglo | Código | Tests |
|---|---|---|
| RN7: el voto de alguien que se da de baja no cuenta ni se muestra | `VotosRevelados` (dominio) y `votosQueCuentan` (`RevelarPoker`) | `TestVotosRevelados`, `TestRevelarPoker_CerrarConError` («solo votó una integrante dada de baja»), escenario «El voto de alguien que se dio de baja no se muestra» |
| Las otras pestañas se recargan al revelar | `recargaSiSeRevelo` en `internal/server/poker_revelar.go`, que usa `quienVoto` de HU-18 | `TestRevelar_FragmentoDeQuienVoto_*` |
