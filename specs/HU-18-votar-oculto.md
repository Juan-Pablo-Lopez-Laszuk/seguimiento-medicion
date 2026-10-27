# SPEC-18 · Votar de forma oculta

**Historia:** HU-18 (#26) · **Responsable:** Piastrellini, Mariano · **Sprint:** 4 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Construye sobre HU-17 (#432): el modelo común de Planning Poker
> (`specs/HU-17-abrir-sesion-poker.md`), la página de la sesión, `quienParticipa` y la tabla `votos` de la migración 0001
> (un voto por integrante y por ronda, con la escala como `CHECK`). Tareas: #275 a #282.
>
> HU-19 (Carolina) revela los votos sobre esta base: la sección **Para HU-19** (al final) dice qué deja esta historia.

## 1. Objetivo

Que cada integrante emita su estimación de la historia sin que los demás la vean hasta que se revele la ronda, para
evitar el efecto ancla.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto e historia | id y número HU-N (vienen en la URL) | Sí | Tienen que existir, y la historia tiene que tener una sesión abierta |
| Quién vota | el integrante | Sí | Con login: el vinculado a la cuenta (TEC-05). Sin login: se elige en el formulario (`quien`). Tiene que ser un integrante activo del proyecto (RN2) |
| Valor | número | Sí | Uno de la escala de HU-16: 0, 1, 2, 3, 5, 8, 13, 21 (CA-18.4) |

## 3. Salidas esperadas

En la página de la sesión (`/proyectos/{id}/backlog/{numero}/poker`), con la sesión abierta y la ronda sin revelar
(bloque `pokerVotar`, llamado desde `pokerRonda`):

- **Las cartas:** un botón por cada valor de la escala. Tocar una carta vota ese valor (`POST …/poker/votos`). Sin
  login, antes se elige **Quién vota** entre los participantes.
- **Quién votó:** cada participante con **Votó** o **Falta votar**, y la cuenta "2 de 3 votaron". **Nunca el valor**
  (CA-18.2), ni siquiera el propio.
- Al votar vuelve a la página con **"Voto registrado. Podés cambiarlo hasta que se revele la ronda."** (303).
- La lista de quién votó se actualiza sola cada 5 segundos con HTMX (`GET …/poker/votos`, que devuelve solo ese
  fragmento). El resto de la página se actualiza al recargar.
  - Sin sesión abierta, o con la ronda revelada, el fragmento queda vacío: HTMX lo saca y deja de pedirlo.
  - Un `GET …/poker/votos` que no viene de HTMX (por ejemplo, recargar después de un error al votar, que queda en esa
    dirección) vuelve a la página de la sesión (303).
- Si no se puede votar, la página de la sesión se vuelve a mostrar con el aviso, o con el error debajo de su campo (las
  cartas o **Quién vota**), y el código de la sección 7.

## 4. Reglas de negocio

- RN1: **Un voto por integrante y por ronda** (CA-18.1). Votar de nuevo en la misma ronda **reemplaza** el voto: en la
  base es un `UPDATE` sobre el índice único `(ronda_id, integrante_id)`.
- RN2: **Votan los participantes:** los integrantes activos del proyecto (CA-17.3), también el Agile Enabler. Uno dado
  de baja no vota, y si ya había votado su voto no cuenta ni se muestra.
- RN3: **Se vota en la ronda actual** y solo mientras no se reveló (HU-19). Después de revelar, votar es un error; el
  voto no se guarda aunque la ronda se revele entre que se abre la página y se vota.
- RN4: **El valor es de la escala** (CA-18.4): `backlog.EscalaStoryPoints`, la misma de HU-16. Cualquier otro valor
  (4, -1, "abc", vacío) es un error y no se guarda nada.
- RN5: **Los valores no salen del servidor antes de revelar** (CA-18.2 y CA-18.3): ni la página, ni el fragmento de
  HTMX, ni la vista del caso de uso los tienen. El repositorio de HU-18 solo devuelve **quiénes** votaron; leer los
  valores lo agrega HU-19, solo para una ronda revelada.
- RN6: **Quién es quién**, como en HU-17 (RN6): con login, la cuenta; sin login, el formulario.

## 5. Restricciones

- **Sin migración:** la tabla `votos` y sus restricciones existen desde la 0001.
- **Dominio puro** en `internal/domain/poker/votar.go`: `ValidarVoto` (RN4), `PuedeVotar` (RN2 y RN3) y `Votaron` (quién
  votó entre los participantes, RN2).
- **Repositorio** en `poker_votos.go` (memoria y PostgreSQL). En PostgreSQL, guardar el voto y controlar que la ronda no
  esté revelada ni la sesión cerrada va en una sola sentencia (`INSERT … SELECT … FOR SHARE … ON CONFLICT DO UPDATE`),
  que bloquea la ronda mientras guarda: si se revela a la vez, el voto no se guarda (RN3).
- **Caso de uso** en `internal/service/poker_votar.go`: `Poker.Votar` y `Poker.ConVotos`, que conecta el repositorio de
  votos sin cambiar `NuevoPoker` (como `Criterios.ConHistorial` en HU-16b). `VistaSesion` suma `Votaron`, el conjunto de
  IDs de los participantes que votaron.
- **Pantalla** con los componentes de TEC-07 y HTMX 2.0.4, que ya carga la base. Sin JavaScript propio.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Valores de los extremos de la escala: 0 y 21 | Se aceptan |
| 4, -1, 100, "abc", "5.5" o vacío | Error de escala (CA-18.4) |
| Votar dos veces | Queda solo el último voto (CA-18.1); la cuenta no cambia |
| Votar el mismo valor dos veces | No es un error: queda igual |
| Nadie votó | "0 de 3 votaron" y todos con Falta votar |
| Todos votaron | "3 de 3 votaron". Revelar es HU-19 |
| Un integrante vota y después se da de baja | Deja de aparecer y su voto no cuenta (RN2) |
| Un integrante nuevo con la ronda en curso | Aparece con Falta votar (RN5 de HU-17) |
| La historia no tiene sesión abierta | Error: no se puede votar |
| La ronda se revela antes de que llegue el voto | Error (RN3) y el voto no se guarda |
| La sesión se cierra antes de que llegue el voto (HU-22) | Error y el voto no se guarda |
| Recargar después de un error al votar | Vuelve a la página de la sesión |

## 7. Condiciones de error

| Situación | Error | Código | Mensaje al usuario |
|---|---|---|---|
| Valor fuera de la escala o vacío | `poker.ErrVotoFueraDeEscala` | 422 | "el voto tiene que ser un valor de la escala: 0, 1, 2, 3, 5, 8, 13 o 21" |
| Quien vota no es un integrante activo | `poker.ErrSoloParticipantes` | 403 | "solo votan los integrantes activos del proyecto" |
| Sin login, no se eligió quién vota | `poker.ErrSoloParticipantes` | 422 | El mismo, debajo del desplegable |
| Con login, la cuenta no es integrante del proyecto | `member.ErrNoEsIntegrante` | 403 | "No sos integrante de este proyecto." |
| La historia no tiene sesión abierta | `poker.ErrSesionNoEncontrada` | 409 | "esta historia no tiene una sesión de Planning Poker abierta" |
| La ronda ya se reveló | `poker.ErrRondaRevelada` | 409 | "la ronda ya se reveló: no se puede votar" |
| La sesión se cerró (HU-22) | `poker.ErrSesionCerrada` | 409 | "la sesión de Planning Poker ya está cerrada" |
| El proyecto o la historia no existen | — | 404 | Los de siempre |
| Falla la base al votar | — | 500 | "No se pudo guardar el voto. Probá de nuevo en unos minutos." |
| Falla la base al ver quién votó | — | 500 | "No se pudo cargar la sesión de Planning Poker. Probá de nuevo en unos minutos." (el de HU-17) |

## 8. Criterios de aceptación (issue #26)

- CA-18.1: Un voto por integrante por ronda; se puede cambiar antes del cierre.
- CA-18.2: Los demás sólo ven quién votó, nunca el valor.
- CA-18.3: La API tampoco expone los votos antes de revelar.
- CA-18.4: Voto fuera de escala → error.

## Decisiones (confirmadas por Mariano el 27/10)

1. **Nadie ve ningún valor antes de revelar, ni el propio** (RN5). Sin login cualquiera puede elegir a otro integrante
   en el formulario, así que mostrar "tu voto" filtraría valores ajenos; con una sola regla no hay casos especiales.
2. **Votar de nuevo reemplaza el voto** hasta que se revele la ronda (RN1 y RN3).
3. **Vota cualquier integrante activo**, incluido el Agile Enabler. El voto de uno que se da de baja deja de contar
   (RN2), así "todos votaron" en HU-19 se cuenta sobre los participantes de ese momento.
4. **El repositorio de HU-18 no devuelve valores:** solo quiénes votaron (RN5). Así CA-18.3 queda garantizado por
   construcción, y HU-19 agrega la lectura de los valores para una ronda revelada.
5. **Cartas como botones** (un formulario, un botón por valor) y **la lista de quién votó se actualiza cada 5 s con
   HTMX** (lo pide la tarea T5, #279). Es solo ese fragmento y sin valores; el resto, al recargar (decisión 1 de la
   Planning).
6. **`Poker.ConVotos`** conecta los votos sin cambiar `NuevoPoker`, así HU-19 y HU-22 pueden hacer lo mismo con lo suyo
   sin chocar.
7. **Un PR con la spec y la feature primero** (hoy), encima de #432, para que Carolina escriba HU-19 sobre esta base; el
   código va en el mismo PR después, con TDD.

## Para HU-19

- **En memoria**, `memory.Poker` ya guarda los votos (`votos`) junto con las sesiones y las rondas: `Revelar` cambia
  `Revelada` en `r.rondas`, y desde entonces `Votar` devuelve `poker.ErrRondaRevelada` (hoy los tests internos de
  `poker_votos_internal_test.go` lo simulan cambiando el campo).
- `ronda.Revelada` existe desde la 0001; HU-19 agrega `Revelar` (repositorio) y la lectura de los votos de una ronda
  **revelada** (por ejemplo `VotosRevelados(ctx, rondaID)`, que devuelve `poker.ErrRondaNoRevelada` si no lo está).
- "Todos votaron" (CA-19.1) = `len(Votaron) == len(Participantes)` de la `VistaSesion`.
- Después de revelar, `PuedeVotar` ya devuelve `ErrRondaRevelada` y la página deja de mostrar las cartas: el bloque
  `pokerVotar` solo se muestra con la ronda sin revelar.

## Trazabilidad

Issue: #26 · Feature: `features/hu-18-votar-oculto.feature` · Pasos BDD: `features/hu18_steps_test.go`

| Tarea | Código | Tests |
|---|---|---|
| T1 (#275) | Esta spec y la feature | — |
| T2 (#276) | `internal/domain/poker/votar.go` (`ValidarVoto`, `PuedeVotar`, `Votaron`) | `internal/domain/poker/votar_test.go` |
| T3 (#277) | `internal/store/memory/poker_votos.go` y `internal/store/postgres/poker_votos.go` (`Votar`, `QuienesVotaron`) | `internal/store/memory/poker_votos_test.go`, `poker_votos_internal_test.go` e `internal/store/postgres/poker_votos_test.go` |
| T4 (#278) | `internal/service/poker_votar.go` (`RepositorioVotosPoker`, `Poker.ConVotos`, `Poker.Votar`; `VistaSesion.Votaron`) | `internal/service/poker_votar_test.go` |
| T5 (#279) | `internal/server/poker_votar.go` (`votar`, `quienVoto`) y los bloques `pokerVotar` y `pokerQuienVoto` de `poker.html` | `internal/server/poker_votar_test.go` |
| T6 (#280) | Que no se expongan los votos (RN5) | `TestVistaSesion_NoTieneValoresDeVotos`, `TestPoker_NadieVeLosValoresAntesDeRevelar` y el escenario «Nadie ve los valores antes de revelar» |
| T7 (#281) | `app/app.go` y `features/hu18_steps_test.go` | `app/poker_test.go` y los escenarios |
| T8 (#282) | Esta spec, la matriz de trazabilidad y el registro de IA | — |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-18.1 Un voto por integrante y por ronda; se cambia antes de revelar | Un integrante vota · Cambiar el voto antes de revelar | `TestPoker_VotarDeNuevo_ReemplazaElVoto` (memoria y PostgreSQL), `TestVotarPoker_DeNuevoYElAgileEnabler`, `TestPoker_VotarSinLogin` |
| CA-18.2 Solo se ve quién votó | Un integrante vota · Nadie ve los valores antes de revelar | `TestVotaron`, `TestVotarPoker_GuardaElVotoYMuestraQuienVoto`, `TestPoker_FragmentoDeQuienVoto`, `TestPoker_NadieVeLosValoresAntesDeRevelar` |
| CA-18.3 La API no expone los votos | Nadie ve los valores antes de revelar | `TestVistaSesion_NoTieneValoresDeVotos`, `TestPoker_NadieVeLosValoresAntesDeRevelar`, `TestPoker_VotarYQuienesVotaron` |
| CA-18.4 Fuera de la escala, error | Valores de la escala · Voto fuera de la escala | `TestValidarVoto`, `TestValidarVoto_FueraDeEscala`, `TestVotarPoker_Errores`, `TestPoker_VotarSinLogin_Errores` |
| RN2 Votan los participantes | Votan todos los participantes · Una integrante dada de baja no vota · El voto de un integrante que se da de baja no cuenta | `TestPuedeVotar`, `TestVotarPoker_IntegranteDeOtroProyecto`, `TestVotarPoker_ElVotoDeQuienSeDaDeBajaNoCuenta`, `TestPoker_ConLogin_VotaElIntegranteDeLaCuenta` |
| RN3 Solo en la ronda actual, antes de revelar | No se vota en una historia sin sesión abierta | `TestPoker_Votar_RondaRevelada` (memoria y PostgreSQL), `TestPoker_Votar_SesionCerrada`, `TestVotarPoker_SeRevelaAntesDeGuardar`, `TestPoker_RondaRevelada` |
