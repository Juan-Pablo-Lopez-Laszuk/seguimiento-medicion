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
