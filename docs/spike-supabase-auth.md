# Spike · Supabase Auth + validación de JWT en Go (Sprint 0)

Investigación para TEC-05 (Autenticación, Sprint 2). No es código de producción: es la base
para no arrancar de cero cuando llegue esa historia.

## Cómo funciona el login (confirma lo que dice la Guía, sección 6.2)

1. El navegador manda email + contraseña a nuestro backend Go.
2. Nuestro backend llama a la API de Supabase Auth (`POST /auth/v1/token?grant_type=password`).
3. Supabase devuelve un **JWT** (access token) + un refresh token.
4. Guardamos el JWT en una cookie `HttpOnly` (no accesible desde JS del navegador, evita XSS).
5. En cada pedido posterior, un **middleware** de Go valida el JWT sin volver a llamar a
   Supabase (eso es justo lo que hay que resolver: cómo validar localmente).

## El punto importante: cómo se valida el JWT sin llamar a Supabase en cada request

Supabase firma los JWT y expone las claves públicas para verificarlos localmente:

- **Endpoint JWKS**: `https://<project-ref>.supabase.co/auth/v1/.well-known/jwks.json`
  (cacheado 10 minutos por Supabase — importante para no rechazar tokens válidos por caché
  vencida si rotan las claves).
- Supabase recomienda algoritmos **asimétricos** (RS256 o ES256) en vez de HS256 (secreto
  compartido).
- Cada JWT trae un header `kid` (key ID). Para validar: se busca en el JWKS la clave pública
  cuyo `kid` coincide, y con esa se verifica la firma.
- **Confirmado (29/09/2026)**: revisé *Project Settings → JWT Keys* en dev y prod, y ambos
  proyectos usan el mismo esquema:
  - **CURRENT KEY**: `ECC (P-256)`, es decir **ES256** (asimétrico) — es la que firma los
    tokens nuevos. Confirma que el camino JWKS de este documento es el correcto.
  - **PREVIOUS KEY**: `Legacy HS256 (Shared Secret)` — Supabase la deja solo para poder
    verificar tokens ya emitidos con el esquema viejo antes de rotar; no firma nada nuevo, así
    que no la vamos a necesitar.

## Claims útiles del token

| Claim | Contenido | Para qué lo usamos |
|---|---|---|
| `sub` | ID único del usuario en Supabase Auth | Lo guardamos como `auth_user_id` en `Integrante` (ver `docs/modelo-datos.md`) |
| `email` | Email del usuario | Para mostrarlo y para matchear con `Integrante.email` |
| `role` | Rol de Postgres (para RLS) | No lo vamos a usar directo; nuestra autorización es a nivel de aplicación (rol Scrum), no de RLS |
| `exp` | Expiración | El middleware debe rechazar tokens vencidos |

## Librerías Go propuestas

- [`github.com/golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt/v5) — parsea y valida el JWT.
- [`github.com/MicahParks/keyfunc/v3`](https://github.com/MicahParks/keyfunc) — resuelve automáticamente
  la clave correcta desde un endpoint JWKS (maneja el matching por `kid` y el refresco de caché).
  Es la combinación estándar en el ecosistema Go para este caso.

## Boceto de middleware (borrador, NO ejecutar todavía — no hay `go.mod`)

```go
// internal/http/middleware/auth.go (Sprint 2, TEC-05)
package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/MicahParks/keyfunc/v3"
    "github.com/golang-jwt/jwt/v5"
)

type claveContexto string

const ClaveUsuarioID claveContexto = "usuarioID"

// RequiereLogin valida el JWT de la cookie y, si es válido, agrega el ID del usuario al contexto.
// jwks se crea una sola vez al arrancar el servidor (apunta al endpoint JWKS del proyecto).
func RequiereLogin(jwks keyfunc.Keyfunc) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            cookie, err := r.Cookie("sesion")
            if err != nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
            }

            token, err := jwt.Parse(cookie.Value, jwks.Keyfunc)
            if err != nil || !token.Valid {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
            }

            claims, _ := token.Claims.(jwt.MapClaims)
            sub, _ := claims["sub"].(string)

            ctx := context.WithValue(r.Context(), ClaveUsuarioID, sub)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

Esto todavía no resuelve el paso "buscar el Integrante a partir del `sub`" (punto 4 de la
Guía) — eso depende de que exista la tabla `Integrante` con `auth_user_id` (ver
`docs/modelo-datos.md`), que se define en el Sprint 0/1.

## Pendientes para cuando arranque TEC-05 (Sprint 2)

1. Decidir con el equipo si armamos el `Keyfunc` una sola vez al iniciar el servidor o por
   request (por el modelo serverless de Vercel, probablemente conviene inicializarlo en el
   handler de `api/index.go` y reusarlo entre invocaciones "calientes").
2. Escribir el primer test (RED) para el middleware usando un JWT de prueba firmado con una
   clave propia (no hace falta pegarle a Supabase real en el test unitario).

## Fuentes

- [JSON Web Token (JWT) | Supabase Docs](https://supabase.com/docs/guides/auth/jwts)
- [JWT Signing Keys | Supabase Docs](https://supabase.com/docs/guides/auth/signing-keys)
- [Introducing JWT Signing Keys (blog)](https://supabase.com/blog/jwt-signing-keys)
