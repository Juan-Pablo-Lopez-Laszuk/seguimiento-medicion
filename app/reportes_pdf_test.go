package app_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// verificarPDFs revisa que los PDF del Sprint 1 y del proyecto se descarguen con su nombre y la fecha de hoy (HU-39).
func verificarPDFs(t *testing.T, h http.Handler) {
	t.Helper()
	hoy := domain.HoyEnArgentina(time.Now()).Format(time.DateOnly)
	for ruta, archivo := range map[string]string{
		"/proyectos/1/sprints/1/reporte.pdf": "con-cierre-sprint-1-" + hoy + ".pdf",
		"/proyectos/1/reporte.pdf":           "con-cierre-" + hoy + ".pdf",
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
		if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Body.String(), "%PDF-") {
			t.Errorf("%s: se esperaba un PDF y se obtuvo %d", ruta, rec.Code)
		}
		if cd := rec.Header().Get("Content-Disposition"); cd != `attachment; filename="`+archivo+`"` {
			t.Errorf("%s: Content-Disposition = %q, se esperaba el archivo %s", ruta, cd, archivo)
		}
	}
}

// HU-39: sin base, los PDF se descargan en memoria.
func TestNewHandlerConBase_SinURLSeDescarganLosPDF(t *testing.T) {
	h := nuevaApp(t, "")
	conSprintActivoHoy(t, h)
	if rec := post(h, "/proyectos/1/sprints/1/cerrar", nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("cerrar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	verificarPDFs(t, h)
}

// HU-39: con base, los PDF se arman con lo que está en PostgreSQL después de un reinicio.
func TestNewHandlerConBase_ConURLLosPDFSobrevivenAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	conSprintActivoHoy(t, primera)
	if rec := post(primera, "/proyectos/1/sprints/1/cerrar", nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("cerrar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	verificarPDFs(t, nuevaApp(t, url))
}
