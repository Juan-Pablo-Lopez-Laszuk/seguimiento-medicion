// Package web contiene los archivos de la interfaz (plantillas HTML) embebidos en el binario de Go,
// así el servidor local y la función de Vercel no dependen de archivos sueltos en disco.
package web

import "embed"

// Templates tiene el layout base (templates/base.html) y una plantilla por página (templates/paginas/).
//
//go:embed templates
var Templates embed.FS
