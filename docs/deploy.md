# Deploy en Vercel (TEC-04)

La app se publica en Vercel con la integración de GitHub. Spec: `specs/TEC-04-deploy-vercel.md`.

## Cómo se publica

| Qué pasa en GitHub | Qué hace Vercel | Dónde se ve |
|---|---|---|
| Merge a `main` | Publica en **producción** | https://seguimiento-medicion.vercel.app |
| Push a una rama con PR abierto | Arma un **preview** y deja el link en el PR | `seguimiento-medicion-git-<rama>-<equipo>.vercel.app` |

Vercel solo sube lo que habilita `.vercelignore` (`api/`, `app/`, `internal/`, `migrations/`, `web/`, `go.mod` y
`go.sum`). Si un paquete nuevo importa otra carpeta, hay que agregarla ahí o el deploy no compila.

## Variables de entorno

Se cargan en Vercel → Settings → Environment Variables. Las carga el dueño del proyecto de Vercel (Juan Pablo).

| Variable | Production | Preview | Para qué |
|---|---|---|---|
| `DATABASE_URL` | Sí (base de producción) | Opcional (base de desarrollo) | Si está, la app guarda en PostgreSQL; si no, en memoria (TEC-03) |
| `SUPABASE_URL` | Sí | **No** | URL del proyecto de Supabase (`https://<ref>.supabase.co`). Con la siguiente, activa el login (TEC-05) |
| `SUPABASE_ANON_KEY` | Sí | **No** | Clave pública *anon* del mismo proyecto (Project Settings → API) |

Los valores nunca se suben a GitHub (`CONTRIBUTING.md` §6). Antes de cargar la de producción hay que aplicar las
migraciones en esa base (`go run ./cmd/migrar up`, ver `CONTRIBUTING.md` §7).

## Cómo comprobar que quedó bien: `/health`

Abrir `/health` en el deploy que se quiere revisar:

| Respuesta | Qué significa | Qué hacer |
|---|---|---|
| `{"status":"ok","base":"ok"}` (200) | La app anda y la base responde | Nada |
| `{"status":"ok","base":"memoria"}` (200) | Al deploy le falta `DATABASE_URL`: lo que se cargue se pierde | Cargar la variable y volver a publicar |
| `{"status":"error","base":"sin conexión"}` (503) | La variable está, pero la base no responde en 3 segundos | Revisar la URL (usuario, clave, puerto 6543) y que la base esté activa en Supabase; el detalle del error está en los logs de Vercel |

Después de cambiar una variable, Vercel no la toma hasta el próximo deploy: hay que hacer **Redeploy** del último.

## Login (TEC-05)

El login se activa solo con las variables: **las dos o ninguna**. Spec: `specs/TEC-05-autenticacion.md`.

| Variables | Qué hace la app |
|---|---|
| Ninguna | No pide login: funciona abierta, como hasta ahora. Es lo que pasa en los previews y en el CI |
| `SUPABASE_URL` y `SUPABASE_ANON_KEY` | Todas las páginas piden sesión y responden 401 sin una válida; `/health`, `/login` y `/registro` quedan abiertas |
| Solo una | **No arranca** y el log dice cuál falta (`falta SUPABASE_URL` o `falta SUPABASE_ANON_KEY`): así un login a medio configurar no queda abierto sin avisar |

Cómo activarlo en producción (lo hace el dueño del proyecto de Vercel, el mismo día que se mergea TEC-05):

1. En Supabase → Authentication, en el proyecto de **producción**: *Allow new users to sign up* activado, *JWT expiry* en
   `28800` segundos (8 horas) y *Confirm email* según lo que decida el equipo.
2. En Vercel cargar `SUPABASE_URL` y `SUPABASE_ANON_KEY` **solo en Production** (no en Preview).
3. Hacer **Redeploy** del último deploy de producción.
4. Comprobar: `/login` abre el formulario, `/proyectos` sin sesión responde 401 y `/health` sigue en 200.

Para desactivarlo, borrar las dos variables y hacer Redeploy. Si después del cambio nadie puede entrar, `/health` no sirve
para saberlo (siempre está abierto): mirar los logs de Vercel, donde se anotan los motivos (por ejemplo, si no se pudieron
bajar las claves de `<SUPABASE_URL>/auth/v1/.well-known/jwks.json`).
