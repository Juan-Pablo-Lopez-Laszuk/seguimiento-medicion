package reportepdf

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// NombreArchivo devuelve el nombre del PDF: proyecto-sprint-fecha.pdf (CA-39.1, RN3). El proyecto va en minúsculas,
// sin tildes y con guiones; sprint 0 es el reporte del proyecto, que no lleva "sprint-N"; la fecha es la de hoy en la
// Argentina, como AAAA-MM-DD.
func NombreArchivo(proyecto string, sprint int, ahora time.Time) string {
	partes := []string{slug(proyecto)}
	if sprint > 0 {
		partes = append(partes, fmt.Sprintf("sprint-%d", sprint))
	}
	partes = append(partes, domain.HoyEnArgentina(ahora).Format("2006-01-02"))
	return strings.Join(partes, "-") + ".pdf"
}

// slug deja solo letras sin tilde y números, separados por un guion: "Gestión Ágil 2026" → "gestion-agil-2026". Si no
// queda nada, devuelve "proyecto".
func slug(texto string) string {
	var b strings.Builder
	guion := false
	for _, r := range norm.NFD.String(strings.ToLower(texto)) {
		switch {
		case unicode.Is(unicode.Mn, r): // la tilde que NFD separó de su letra
		case r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			if guion && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			guion = false
		default:
			guion = true
		}
	}
	if b.Len() == 0 {
		return "proyecto"
	}
	return b.String()
}
