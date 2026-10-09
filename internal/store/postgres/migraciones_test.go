package postgres_test

import (
	"context"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// tablas del modelo acordado (docs/modelo-datos.md).
var tablas = []string{
	"proyectos", "integrantes", "sprints", "historias", "criterios_aceptacion", "tareas",
	"sprint_historias", "sesiones_poker", "rondas_poker", "votos", "registros_esfuerzo", "defectos",
}

func TestMigracion_SubeTodasLasTablasConRLS(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)

	var total, conRLS int
	err := pool.QueryRow(ctx, `SELECT count(*), count(*) FILTER (WHERE rowsecurity)
		FROM pg_tables WHERE schemaname = current_schema() AND tablename <> 'goose_db_version'`).Scan(&total, &conRLS)
	if err != nil {
		t.Fatalf("consultar las tablas: %v", err)
	}
	if total != len(tablas) {
		t.Errorf("se esperaban %d tablas y hay %d", len(tablas), total)
	}
	if conRLS != total {
		t.Errorf("todas las tablas deben tener RLS activado: %d de %d", conRLS, total)
	}
	for _, tabla := range tablas {
		var existe bool
		if err := pool.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, tabla).Scan(&existe); err != nil || !existe {
			t.Errorf("falta la tabla %s (err=%v)", tabla, err)
		}
	}
}

func TestMigracion_BajaDejaLaBaseSinTablasDelModelo(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)

	if err := postgres.Bajar(ctx, pool); err != nil {
		t.Fatalf("bajar las migraciones: %v", err)
	}
	var restantes int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM pg_tables
		WHERE schemaname = current_schema() AND tablename <> 'goose_db_version'`).Scan(&restantes); err != nil {
		t.Fatalf("consultar las tablas: %v", err)
	}
	if restantes != 0 {
		t.Errorf("después de bajar no debería quedar ninguna tabla y quedan %d", restantes)
	}
	// y se puede volver a subir
	if err := postgres.Subir(ctx, pool); err != nil {
		t.Errorf("volver a subir las migraciones: %v", err)
	}
}

func TestMigracion_SubirDosVecesNoFalla(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	if err := postgres.Subir(ctx, pool); err != nil {
		t.Errorf("la segunda ejecución no debería hacer nada ni fallar: %v", err)
	}
}

// Cada fila es una sentencia que debe pasar (ok) o ser rechazada por la base (!ok). Van en orden porque
// las últimas usan las filas que crean las primeras: son las restricciones de docs/modelo-datos.md.
func TestMigracion_RestriccionesDelModelo(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)

	casos := []struct {
		nombre string
		ok     bool
		sql    string
	}{
		{"proyecto válido", true, `INSERT INTO proyectos (nombre, fecha_inicio, fecha_fin, estado) VALUES ('Alfa', '2026-09-28', '2026-11-01', 'Planificado')`},
		{"nombre repetido con otras mayúsculas", false, `INSERT INTO proyectos (nombre, fecha_inicio, fecha_fin, estado) VALUES ('ALFA', '2026-09-28', '2026-11-01', 'Planificado')`},
		{"fecha de fin igual a la de inicio", false, `INSERT INTO proyectos (nombre, fecha_inicio, fecha_fin, estado) VALUES ('Beta', '2026-09-28', '2026-09-28', 'Planificado')`},
		{"sprint válido", true, `INSERT INTO sprints (proyecto_id, numero, objetivo, fecha_inicio, fecha_fin, estado) VALUES (1, 1, 'Meta', '2026-10-05', '2026-10-11', 'Planificado')`},
		{"número de sprint repetido en el proyecto", false, `INSERT INTO sprints (proyecto_id, numero, objetivo, fecha_inicio, fecha_fin, estado) VALUES (1, 1, 'Otra', '2026-10-12', '2026-10-18', 'Planificado')`},
		{"historia válida", true, `INSERT INTO historias (proyecto_id, numero, titulo, prioridad, story_points, horas_estimadas) VALUES (1, 1, 'HU-05', 'Alta', 5, 1.5)`},
		{"historia sin estimar (SP nulo)", true, `INSERT INTO historias (proyecto_id, numero, titulo, prioridad) VALUES (1, 2, 'HU-06', 'Media')`},
		{"SP 21, el máximo de la escala", true, `INSERT INTO historias (proyecto_id, numero, titulo, prioridad, story_points) VALUES (1, 3, 'x', 'Alta', 21)`},
		{"SP 4, fuera de la escala", false, `INSERT INTO historias (proyecto_id, numero, titulo, prioridad, story_points) VALUES (1, 4, 'x', 'Alta', 4)`},
		{"SP 34, fuera de la escala", false, `INSERT INTO historias (proyecto_id, numero, titulo, prioridad, story_points) VALUES (1, 4, 'x', 'Alta', 34)`},
		{"prioridad inválida", false, `INSERT INTO historias (proyecto_id, numero, titulo, prioridad) VALUES (1, 4, 'x', 'Urgente')`},
		{"estado de historia inválido", false, `INSERT INTO historias (proyecto_id, numero, titulo, prioridad, estado) VALUES (1, 4, 'x', 'Alta', 'Terminada')`},
		{"horas estimadas en 0", false, `INSERT INTO historias (proyecto_id, numero, titulo, prioridad, horas_estimadas) VALUES (1, 4, 'x', 'Alta', 0)`},
		{"número de historia repetido en el proyecto", false, `INSERT INTO historias (proyecto_id, numero, titulo, prioridad) VALUES (1, 1, 'x', 'Alta')`},
		{"primer Agile Enabler", true, `INSERT INTO integrantes (proyecto_id, nombre, email, rol) VALUES (1, 'JP', 'jp@x.com', 'AgileEnabler')`},
		{"segundo Agile Enabler en el mismo proyecto", false, `INSERT INTO integrantes (proyecto_id, nombre, email, rol) VALUES (1, 'Otro', 'o@x.com', 'AgileEnabler')`},
		{"Product Builder", true, `INSERT INTO integrantes (proyecto_id, nombre, email, rol) VALUES (1, 'Mariano', 'm@x.com', 'ProductBuilder')`},
		{"email repetido en el proyecto", false, `INSERT INTO integrantes (proyecto_id, nombre, email, rol) VALUES (1, 'Dup', 'm@x.com', 'ProductBuilder')`},
		{"sesión de poker abierta", true, `INSERT INTO sesiones_poker (historia_id, estado) VALUES (1, 'Abierta')`},
		{"segunda sesión abierta de la misma historia", false, `INSERT INTO sesiones_poker (historia_id, estado) VALUES (1, 'Abierta')`},
		{"sesión cerrada con valor 8", true, `INSERT INTO sesiones_poker (historia_id, estado, valor_acordado) VALUES (1, 'Cerrada', 8)`},
		{"valor acordado fuera de la escala", false, `INSERT INTO sesiones_poker (historia_id, estado, valor_acordado) VALUES (1, 'Cerrada', 4)`},
		{"ronda 1", true, `INSERT INTO rondas_poker (sesion_id, numero) VALUES (1, 1)`},
		{"ronda 1 repetida en la sesión", false, `INSERT INTO rondas_poker (sesion_id, numero) VALUES (1, 1)`},
		{"voto válido", true, `INSERT INTO votos (ronda_id, integrante_id, valor) VALUES (1, 1, 5)`},
		{"segundo voto del mismo integrante en la ronda", false, `INSERT INTO votos (ronda_id, integrante_id, valor) VALUES (1, 1, 8)`},
		{"voto fuera de la escala", false, `INSERT INTO votos (ronda_id, integrante_id, valor) VALUES (1, 2, 4)`},
		{"esfuerzo de 8 horas", true, `INSERT INTO registros_esfuerzo (integrante_id, historia_id, fecha, horas) VALUES (1, 1, '2026-10-09', 8)`},
		{"esfuerzo de 25 horas", false, `INSERT INTO registros_esfuerzo (integrante_id, historia_id, fecha, horas) VALUES (1, 1, '2026-10-09', 25)`},
		{"esfuerzo de 0 horas", false, `INSERT INTO registros_esfuerzo (integrante_id, historia_id, fecha, horas) VALUES (1, 1, '2026-10-09', 0)`},
		{"esfuerzo sin historia ni tarea", false, `INSERT INTO registros_esfuerzo (integrante_id, fecha, horas) VALUES (1, '2026-10-09', 2)`},
		{"foto de cierre con SP nulo", true, `INSERT INTO sprint_historias (sprint_id, historia_id, story_points) VALUES (1, 2, NULL)`},
		{"foto de cierre repetida", false, `INSERT INTO sprint_historias (sprint_id, historia_id) VALUES (1, 2)`},
		{"foto de cierre con SP fuera de la escala", false, `INSERT INTO sprint_historias (sprint_id, historia_id, story_points) VALUES (1, 1, 7)`},
		{"tarea con horas negativas", false, `INSERT INTO tareas (historia_id, titulo, horas_estimadas) VALUES (1, 't', -1)`},
		{"defecto válido", true, `INSERT INTO defectos (proyecto_id, descripcion, severidad, estado, sprint_deteccion_id) VALUES (1, 'd', 'Alta', 'Abierto', 1)`},

		// HU-03 (RN6 y RN4): un solo Agile Enabler ACTIVO por proyecto; el email de quien se dio de baja sigue ocupado.
		{"dar de baja al Agile Enabler", true, `UPDATE integrantes SET activo = false WHERE email = 'jp@x.com'`},
		{"nuevo Agile Enabler cuando el anterior está de baja", true, `INSERT INTO integrantes (proyecto_id, nombre, email, rol) VALUES (1, 'Nuevo AE', 'ae2@x.com', 'AgileEnabler')`},
		{"otro Agile Enabler activo más", false, `INSERT INTO integrantes (proyecto_id, nombre, email, rol) VALUES (1, 'Otro AE', 'ae3@x.com', 'AgileEnabler')`},
		{"reactivar al Agile Enabler que estaba de baja", false, `UPDATE integrantes SET activo = true WHERE email = 'jp@x.com'`},
		{"registrar el email de alguien dado de baja", false, `INSERT INTO integrantes (proyecto_id, nombre, email, rol) VALUES (1, 'Repetido', 'jp@x.com', 'ProductBuilder')`},
	}
	for _, c := range casos {
		_, err := pool.Exec(ctx, c.sql)
		if c.ok && err != nil {
			t.Errorf("%s: se esperaba que la base lo aceptara y falló: %v", c.nombre, err)
		}
		if !c.ok && err == nil {
			t.Errorf("%s: se esperaba que la base lo rechazara y lo aceptó", c.nombre)
		}
	}
}
