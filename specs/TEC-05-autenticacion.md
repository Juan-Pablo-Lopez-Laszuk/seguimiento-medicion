# SPEC-TEC-05 · Autenticación de usuarios con Supabase Auth

**Tarea:** TEC-05 (#68) · **Responsable:** Piastrellini, Mariano · **Sprint:** 3 (adelantada del Sprint 4) · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`. Parte del spike
> `docs/spike-supabase-auth.md` (#66) y de la Guía de desarrollo, sección 6.2. Depende de HU-03 (el email de los
> integrantes, normalizado) y de la columna `integrantes.auth_user_id`, que existe desde TEC-03.

## 1. Objetivo

Que cada integrante **se registre e inicie sesión** para que el sistema sepa **quién es quién**: quién carga horas y,
a partir del Sprint 4, quién vota en Planning Poker. El login lo hace Supabase Auth; la aplicación guarda el token en una
cookie, lo **valida en cada pedido sin volver a llamar a Supabase** y **vincula** a cada usuario con sus integrantes.

Esta historia es solo la **identidad**: registrarse, entrar, salir, saber quién es el usuario y a qué integrante
corresponde en cada proyecto. Qué puede hacer cada rol (autorización) y reemplazar los desplegables de "quién" de otras
historias queda fuera (sección 5).

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Email (registro e ingreso) | texto | Sí | Mismo formato que el de un integrante (HU-03 RN2 y RN3); se guarda sin espacios y en minúsculas |
| Contraseña (registro) | texto | Sí | Entre 8 y 72 caracteres, y no más de 72 bytes (el límite de bcrypt, que usa Supabase) |
| Repetir contraseña (registro) | texto | Sí | Igual a la contraseña |
| Contraseña (ingreso) | texto | Sí | No se le exige largo: la valida Supabase |
| Cookie `sesion` | JWT de Supabase | En todas las rutas menos las abiertas | Ver RN2 |
| `siguiente` (query del login) | ruta de la app | No | Solo rutas internas (RN9) |
| `SUPABASE_URL` y `SUPABASE_ANON_KEY` | variables de entorno | No | Las dos o ninguna (RN1) |

## 3. Salidas esperadas

- **`/registro`** (formulario de email, contraseña y repetirla): si sale bien, vuelve a `/login` con el mensaje
  "Cuenta creada. Ya podés iniciar sesión." (o "Te enviamos un email para confirmar la cuenta. Confirmala y después iniciá
  sesión." si Supabase pide confirmar).
- **`/login`** (formulario de email y contraseña): si sale bien, guarda la cookie `sesion` y va (303) a `siguiente` o a `/`.
- **`POST /salir`**: borra la cookie y vuelve a `/login` con "Cerraste sesión.".
- **Barra de navegación**: con sesión muestra el email del usuario y el botón "Salir"; sin login configurado no muestra
  nada nuevo.
- **Pedido sin sesión válida**: respuesta **401** con la página de login y el motivo (RN3).
- **Vínculo**: los integrantes con el email del usuario quedan con su `auth_user_id`; el caso de uso `Quien` devuelve el
  integrante de un proyecto para el usuario que está usando la app.

## 4. Reglas de negocio

- RN1: **Se activa por configuración.** Si no hay `SUPABASE_URL` ni `SUPABASE_ANON_KEY`, la aplicación no pide login y
  funciona como hasta ahora (modo memoria, previews, tests y CI). Si están las dos, **todas las rutas piden sesión**
  menos `/health`, `/login` y `/registro`. Si está una sola, la aplicación **no arranca** y dice cuál falta (igual que
  con una `DATABASE_URL` inválida): un login a medio configurar no puede quedar abierto sin avisar.
- RN2: **Un token es válido si** está firmado con una clave pública del JWKS del proyecto
  (`SUPABASE_URL/auth/v1/.well-known/jwks.json`, la clave se elige por el `kid` del header), su algoritmo es **ES256 o
  RS256** (se rechazan HS256 y `none`), **no venció**, el emisor (`iss`) es `SUPABASE_URL/auth/v1`, la audiencia (`aud`) es
  `authenticated` y trae un `sub` que es un UUID. Se verifica **localmente**: Supabase solo se consulta para registrar
  e ingresar, y el JWKS se baja una vez y se reutiliza mientras la función siga activa.
- RN3: **Sin sesión válida el pedido se rechaza con 401** y no llega a ningún handler. Vale para la cookie ausente,
  mal formada, con la firma de otra clave, con otro algoritmo, de otro emisor o audiencia, o vencida. La respuesta es la
  página de login (para que el navegador muestre el formulario) con "Iniciá sesión para continuar." o, si el token
  estaba vencido, "Tu sesión venció, iniciá sesión de nuevo.". Una cookie inválida o vencida se borra en esa respuesta.
- RN4: **La cookie** se llama `sesion`, es `HttpOnly` (el JavaScript de la página no la lee), `SameSite=Lax` (los
  formularios de otro sitio no la mandan, lo que cubre el CSRF de los POST), de `Path=/`, `Secure` cuando el pedido llegó
  por HTTPS (incluido el `X-Forwarded-Proto` de Vercel) y vence cuando vence el token. Guarda solo el token de acceso.
  **No hay renovación**: al vencer hay que ingresar de nuevo; cuánto dura lo fija Supabase (*JWT expiry*).
- RN5: **Registro.** Se validan el email, el largo de la contraseña y que las dos coincidan, con todos los errores a la
  vez debajo de cada campo y lo escrito conservado (menos las contraseñas). Después se crea la cuenta en Supabase. Si ese
  email ya tiene cuenta: "ya hay una cuenta con ese email" (con *Confirm email* activado Supabase no lo informa, por
  privacidad: la persona ve el mismo "te enviamos un email"). El registro **es abierto**: cualquier email puede registrarse
  (la spec de HU-01 pide que cualquier usuario con sesión pueda crear un proyecto); estar o no en un proyecto se resuelve
  con el vínculo (RN7).
- RN6: **Ingreso.** Con las credenciales correctas guarda la cookie y vincula (RN7). Con credenciales incorrectas
  responde **401** "El email o la contraseña no son correctos." sin decir cuál de los dos, y no guarda ninguna cookie. Si
  Supabase exige confirmar el email y todavía no se confirmó: "Confirmá tu email antes de iniciar sesión.".
- RN7: **Vínculo con los integrantes** (decisión 1 de `docs/modelo-datos.md`). Al ingresar, el usuario (`sub`) se vincula
  con **todos** los integrantes, de cualquier proyecto, cuyo email es el de su cuenta y que **todavía no tienen un
  usuario**. Un integrante ya vinculado con otro usuario no se pisa; uno ya vinculado con este mismo usuario no cambia. Los
  dados de baja también se vinculan. Un usuario puede quedar vinculado con un integrante por proyecto.
- RN8: **`Quien(proyecto)`** devuelve el integrante del proyecto vinculado con el usuario. Si no hay uno vinculado pero hay
  un integrante del proyecto con su email (lo registraron **después** de que ingresó), lo vincula en ese momento. Si el
  usuario no es integrante del proyecto, el error es `ErrNoEsIntegrante`. Quien lo use decide qué hacer con uno dado de
  baja (viene con `Activo = false`).
- RN9: **Volver a donde se estaba.** `siguiente` solo se acepta si es una ruta interna: empieza con `/` y no con `//` ni
  con `/\`. Cualquier otro valor se ignora y va a `/` (evita que el login redirija a otro sitio).
- RN10: **Salir** borra la cookie en el navegador. No revoca el token en Supabase: como se verifica localmente, un token
  copiado sigue valiendo hasta que vence (por eso conviene que dure poco).
- RN11: **Nada sensible en los logs ni en las pantallas.** No se escriben contraseñas ni tokens; los mensajes de error de
  Supabase no se muestran tal cual; el ingreso no revela si un email existe.
- RN12: **Si el servicio de login no responde** (sin conexión, error 5xx o más de 5 segundos) el registro y el ingreso
  responden **503** "No pudimos conectar con el servicio de inicio de sesión. Probá de nuevo en unos minutos."; si
  Supabase limita los intentos (429): "Demasiados intentos. Esperá un momento y volvé a probar.". Si no se puede bajar el
  JWKS, los pedidos con cookie se rechazan con 401 (falla cerrada) y el motivo va al log.

## 5. Restricciones

- **Dependencias nuevas**, las del spike: `github.com/golang-jwt/jwt/v5` (parsea y valida el JWT) y
  `github.com/MicahParks/keyfunc/v3` (resuelve la clave del JWKS por `kid` y la cachea).
- **Capas.** La validación de email y contraseña es dominio puro (`internal/domain/account`). El token y el cliente de
  Supabase van en `internal/auth`, detrás de interfaces (`Verificador` y `ServicioDeCuentas`), así los tests no usan red:
  firman un JWT ES256 con una clave propia, sirven el JWKS con `httptest` y reemplazan Supabase con un servidor falso.
  El caso de uso `Cuentas` (`Registrar`, `Ingresar`, `Quien`) está en `internal/service`; las pantallas y el middleware
  en `internal/server`; el armado en `app/app.go`.
- **Repositorios de integrantes** (memoria y PostgreSQL): `Integrante` suma `AuthUserID` y los repositorios suman
  `VincularPorEmail` y `BuscarPorUsuario`. **Sin migración nueva**: la columna `auth_user_id uuid` ya existe.
- **Los tests y el CI no necesitan variables ni red**: sin ellas la aplicación no pide login (RN1). Las dos variables
  (los mismos nombres de `.env.example`) se cargan en Vercel solo en **Production**; en Preview no, así los previews de
  los PRs siguen abiertos.
- `/health` siempre queda abierto: lo usan Vercel y el equipo para saber si el deploy anda.
- Se podría guardar quién estimó en el historial de HU-16b o quién votó en Planning Poker, pero eso es de esas historias:
  esta deja disponible `Quien`.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Contraseña de 7 caracteres | Error en el registro: "la contraseña tiene que tener entre 8 y 72 caracteres" |
| Contraseña de 8 caracteres / de 72 | Se acepta |
| Contraseña de 73 caracteres, o de menos pero con más de 72 bytes (por ejemplo, acentos) | Error en el registro (bcrypt corta en 72 bytes) |
| Las contraseñas no coinciden | Error debajo de "Repetir contraseña" |
| Email con mayúsculas y espacios (` Ana@Mail.com `) | Se usa `ana@mail.com` |
| Registro de un email que ya tiene cuenta | "ya hay una cuenta con ese email" |
| Token vencido hace un segundo | 401 "Tu sesión venció…" |
| Token firmado con otra clave, con `alg: none` o HS256 | 401 |
| Token de otro emisor o con otra audiencia | 401 |
| `sub` que no es un UUID | 401 |
| Cookie con basura | 401 y se borra |
| Dos proyectos con el mismo email de integrante | El usuario queda vinculado en los dos (RN7) |
| Integrante ya vinculado con otro usuario | No se pisa |
| Integrante registrado después del ingreso | `Quien` lo vincula en ese momento (RN8) |
| `siguiente=https://malo.com`, `//malo.com` o `/\malo.com` | Se ignora: va a `/` |
| `/health` con el login activo y sin cookie | 200 |
| Sin variables de Supabase | No se pide login (RN1) |

## 7. Condiciones de error

| Situación | Código | Mensaje al usuario |
|---|---|---|
| Sin sesión, token inválido | 401 | "Iniciá sesión para continuar." (página de login) |
| Token vencido | 401 | "Tu sesión venció, iniciá sesión de nuevo." |
| Credenciales incorrectas | 401 | "El email o la contraseña no son correctos." |
| Email sin confirmar | 401 | "Confirmá tu email antes de iniciar sesión." |
| Datos del registro inválidos | 422 | El error de cada campo, debajo del campo |
| Email ya registrado | 422 | "ya hay una cuenta con ese email" |
| Supabase no responde o falla | 503 | "No pudimos conectar con el servicio de inicio de sesión. Probá de nuevo en unos minutos." |
| Demasiados intentos | 429 | "Demasiados intentos. Esperá un momento y volvé a probar." |
| Solo una de las dos variables | no arranca | "falta SUPABASE_URL" o "falta SUPABASE_ANON_KEY" |

## 8. Criterios de aceptación (issue #68)

- CA-05.1: Registro e inicio de sesión con Supabase Auth.
- CA-05.2: El middleware de Go valida el JWT y rechaza con 401 si es inválido o expiró.
- CA-05.3: Cada usuario se vincula con un integrante del proyecto.

## Decisiones a confirmar (antes de programar)

1. **401 con la página de login** en lugar de redirigir con 303 como dice la Guía. El criterio del issue pide 401; la
   página de login en el cuerpo hace que el navegador igual muestre el formulario. *Recomendado.*
2. **Registro abierto** a cualquier email y vínculo por email al ingresar (la spec de HU-01 y la decisión 1 de
   `modelo-datos.md` van en ese sentido). Riesgo aceptado: sin confirmación de email, quien conozca el email de un
   integrante podría registrarse antes que él. Si Supabase tiene *Confirm email* activado, el ingreso exige haberlo
   confirmado y el riesgo desaparece **sin cambiar el código**.
3. **Sin renovación del token** (sin refresh): al vencer se vuelve a ingresar. Se pide a Juan Pablo que suba el *JWT
   expiry* de Supabase a 8 horas (el valor por defecto es 1) para no sacar a nadie a mitad de una sesión de Planning Poker.
4. **Contraseña de 8 a 72 caracteres** (Supabase pide 6 por defecto).
5. **Fuera de alcance, para el Sprint 4 o un PR aparte:** reemplazar los desplegables de "quién" de HU-24 y HU-26 por el
   usuario con sesión; autorizar por rol (cerrar un sprint solo el Agile Enabler, HU-14); que quien crea un proyecto
   quede como su Agile Enabler (HU-01); "olvidé mi contraseña"; cambiar el email o la contraseña.

## Pedido a Juan Pablo (configuración, sin código)

- **Supabase Auth, en el proyecto de producción** (y en el de desarrollo si se prueba ahí): *Allow new users to sign up*
  activado; *Confirm email* según la decisión 2; *JWT expiry* en 28800 segundos (decisión 3); contraseña mínima de 6 o más.
- **Vercel, solo Production:** `SUPABASE_URL` (`https://<ref>.supabase.co`) y `SUPABASE_ANON_KEY` (la clave pública
  *anon/publishable*), **el mismo día del merge**; después un Redeploy. Hasta entonces la app sigue sin login.
- No hace falta cargar nada en el CI ni en Preview.

## Trazabilidad

Issue: #68 · Feature: `features/tec-05-autenticacion.feature` · Pasos BDD: `features/tec05_steps_test.go`

Las columnas de código y tests se completan al implementar.
