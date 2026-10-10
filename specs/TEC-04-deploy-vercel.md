# SPEC-TEC-04 · Deploy en Vercel con chequeo de la base

**Tarea:** TEC-04 (#15) · **Responsable:** Bravo, Carolina · **Sprint:** 1 · **Estado:** Implementada (falta que Juan Pablo cargue `DATABASE_URL` en Vercel)

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`. Depende de TEC-03 (#67, PR #91):
> la conexión a PostgreSQL (`postgres.Conectar`) y la variable `DATABASE_URL`.

## 1. Objetivo

Que la app publicada en Vercel diga si **de verdad puede guardar datos**. Hasta ahora `/health` respondía "ok" aunque
la base no estuviera conectada. Con TEC-03 la app guarda en PostgreSQL si existe `DATABASE_URL`; `/health` tiene que
comprobar esa conexión para que el equipo (y los profes) sepan si el deploy quedó bien configurado.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Pedido `GET /health` | HTTP | Sí | — |
| `DATABASE_URL` (variable de entorno de Vercel) | texto | No | Si está, la app usa PostgreSQL; si no, memoria (TEC-03) |

## 3. Salidas esperadas

Siempre JSON con dos campos, `status` y `base`:

| Situación | Código | Respuesta |
|---|---|---|
| Sin `DATABASE_URL` (modo memoria) | 200 | `{"status":"ok","base":"memoria"}` |
| Con base y la base responde | 200 | `{"status":"ok","base":"ok"}` |
| Con base y la base no responde | 503 | `{"status":"error","base":"sin conexión"}` |

## 4. Reglas de negocio

- RN1: `/health` sigue respondiendo 200 cuando la app corre en memoria (desarrollo, tests y previews sin base), pero
  lo dice en `base`, así se ve enseguida si a un deploy le falta la variable.
- RN2: Si hay base, `/health` le hace un **ping**. Si falla, responde **503** (servicio no disponible): la app está
  viva pero no puede guardar.
- RN3: El ping espera **como máximo 3 segundos**. Si la base tarda más, cuenta como "sin conexión"; `/health` nunca
  se queda colgado.
- RN4: La respuesta **no muestra el error de la base** (puede tener el host o el usuario). El detalle va al log del
  servidor, que solo ve el equipo en Vercel.

## 5. Restricciones

- `internal/server` no conoce PostgreSQL: recibe en `Dependencias` algo que sabe hacer `Ping(ctx) error`
  (`server.VerificadorBase`). El pool de pgx ya cumple esa interfaz, y en los tests se usa una base falsa.
- `app.NewHandlerConBase` pasa el pool como verificador solo si hay URL; sin URL no pasa nada (modo memoria).
- Sin base de datos nueva ni migraciones: solo se usa la conexión de TEC-03.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Sin `DATABASE_URL` | 200, `base: memoria` |
| Base que responde | 200, `base: ok` |
| Base caída o URL que apunta a un puerto cerrado | 503, `base: sin conexión` |
| Base que tarda más de 3 s | 503, `base: sin conexión` (el ping se corta a los 3 s) |
| El error de la base trae el usuario o el host | No aparece en la respuesta |

## 7. Condiciones de error

| Situación | Código | Mensaje al usuario |
|---|---|---|
| La base no responde o tarda más de 3 s | 503 | `{"status":"error","base":"sin conexión"}` |

## 8. Criterios de aceptación (issue #15)

- CA-04.1: Cada merge a `main` publica en producción y cada PR genera un preview. (Ya funciona con la integración de
  GitHub de Vercel desde el Sprint 0; se documenta en `docs/deploy.md`.)
- CA-04.2: Variables de entorno configuradas. (`DATABASE_URL` en Vercel la carga Juan Pablo, dueño del proyecto de
  Vercel; `docs/deploy.md` dice cuáles y cómo comprobarlas con `/health`.)
- CA-04.3: `/health` responde 200 y chequea la conexión a la base.

## Trazabilidad

Issue: #15 · Feature: `features/tec-04-health.feature` · Pasos BDD: `features/tec04_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| HTTP | `internal/server/health.go` (`health`, `verificarBase`, `estadoGeneral`, `VerificadorBase`) | `internal/server/health_test.go` |
| Armado de la app | `app/app.go` (`NewHandlerConBase` pasa el pool como `Base`) | `app/app_test.go` |
| Documentación | `docs/deploy.md` (cómo se publica, variables y cómo leer `/health`) | — |

| Criterio / regla | Escenarios BDD | Tests |
|---|---|---|
| CA-04.3 y RN2 (ping, 503) | La base de datos responde · La base de datos no responde | `TestHealth_ConBaseQueRespondeInformaOK`, `TestHealth_ConBaseCaidaResponde503`, `TestNewHandlerConBase_ConBaseHealthInformaOK`, `TestNewHandlerConBase_ConBaseInalcanzableHealthResponde503` |
| RN1 (memoria) | La app corre sin base de datos | `TestHealth_SinBaseInformaMemoria`, `TestNewHandlerConBase_SinURLHealthInformaMemoria` |
| RN3 (3 segundos) | La base de datos tarda en responder (2 y 4 s) | `TestHealth_EsperaALaBaseComoMaximo3Segundos` |
| RN4 (no mostrar el error) | — | `TestHealth_NoMuestraElErrorDeLaBase` |
| CA-04.1 y CA-04.2 | — | Se comprueban en Vercel con `/health` (`docs/deploy.md`) |
