package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// Los usuarios son los UUID que Supabase Auth le da a cada cuenta (TEC-05); la columna auth_user_id es de tipo uuid.
const (
	usuarioAna  = "11111111-1111-4111-8111-111111111111"
	usuarioOtro = "22222222-2222-4222-8222-222222222222"
)

// CA-05.3 y RN7 · el usuario se vincula con todos los integrantes que tienen su email, de cualquier proyecto y estén o no
// dados de baja, y con ningún otro
func TestIntegrantes_VincularPorEmail_VinculaATodosLosQueTienenElEmail(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, proyectoB := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)
	ana, _ := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Ana", "ana@mail.com", member.RolProductBuilder))
	beto, _ := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Beto", "beto@mail.com", member.RolProductBuilder))
	anaEnB, _ := repo.Guardar(ctx, integranteDeEjemplo(proyectoB, "Ana en B", "ana@mail.com", member.RolProductBuilder))
	anaEnB.Activo = false
	if err := repo.Actualizar(ctx, anaEnB); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	vinculados, err := repo.VincularPorEmail(ctx, "ana@mail.com", usuarioAna)

	if err != nil || vinculados != 2 {
		t.Fatalf("se esperaban 2 integrantes vinculados sin error y se obtuvo %d, %v", vinculados, err)
	}
	if i, err := repo.BuscarPorUsuario(ctx, proyectoA, usuarioAna); err != nil || i.ID != ana.ID {
		t.Errorf("proyecto A: se esperaba a Ana vinculada y se obtuvo %+v, %v", i, err)
	}
	if i, err := repo.BuscarPorUsuario(ctx, proyectoB, usuarioAna); err != nil || i.ID != anaEnB.ID || i.Activo {
		t.Errorf("proyecto B: se esperaba a Ana de baja vinculada y se obtuvo %+v, %v", i, err)
	}
	if i, _ := repo.BuscarPorID(ctx, proyectoA, beto.ID); i.AuthUserID != "" {
		t.Errorf("Beto tiene otro email y no se tendría que vincular: %q", i.AuthUserID)
	}
}

// El email se compara sin distinguir mayúsculas, como el resto de la aplicación
func TestIntegrantes_VincularPorEmail_NoDistingueMayusculas(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)
	_, _ = repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Ana", "ana@mail.com", member.RolProductBuilder))

	vinculados, err := repo.VincularPorEmail(ctx, "Ana@Mail.com", usuarioAna)

	if err != nil || vinculados != 1 {
		t.Errorf("se esperaba 1 integrante vinculado sin error y se obtuvo %d, %v", vinculados, err)
	}
}

// RN7 · un integrante ya vinculado con otro usuario no se pisa, y uno ya vinculado con este usuario no se cuenta de nuevo
func TestIntegrantes_VincularPorEmail_NoPisaAUnIntegranteYaVinculado(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)
	_, _ = repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Ana", "ana@mail.com", member.RolProductBuilder))
	if _, err := repo.VincularPorEmail(ctx, "ana@mail.com", usuarioOtro); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	vinculados, err := repo.VincularPorEmail(ctx, "ana@mail.com", usuarioAna)

	if err != nil || vinculados != 0 {
		t.Errorf("se esperaban 0 integrantes vinculados sin error y se obtuvo %d, %v", vinculados, err)
	}
	if i, err := repo.BuscarPorUsuario(ctx, proyectoA, usuarioOtro); err != nil || i.Nombre != "Ana" {
		t.Errorf("Ana tenía que seguir vinculada con el otro usuario y se obtuvo %+v, %v", i, err)
	}
	if _, err := repo.BuscarPorUsuario(ctx, proyectoA, usuarioAna); !errors.Is(err, member.ErrNoEncontrado) {
		t.Errorf("el usuario nuevo no tendría que estar vinculado y se obtuvo %v", err)
	}
	if vinculados, _ := repo.VincularPorEmail(ctx, "ana@mail.com", usuarioOtro); vinculados != 0 {
		t.Errorf("volver a vincular al mismo usuario no cambia nada y se contaron %d", vinculados)
	}
}

func TestIntegrantes_VincularPorEmail_SinCoincidenciasOConUnUsuarioInvalido(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)
	_, _ = repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Ana", "ana@mail.com", member.RolProductBuilder))

	if vinculados, err := repo.VincularPorEmail(ctx, "nadie@mail.com", usuarioAna); err != nil || vinculados != 0 {
		t.Errorf("sin coincidencias se esperaban 0 vinculados sin error y se obtuvo %d, %v", vinculados, err)
	}
	for _, usuario := range []string{"", "no-es-un-uuid"} {
		if _, err := repo.VincularPorEmail(ctx, "ana@mail.com", usuario); err == nil {
			t.Errorf("con el usuario %q se esperaba un error", usuario)
		}
	}
}

// BuscarPorUsuario busca dentro del proyecto: con el de otro proyecto o con un usuario sin vínculo no se encuentra
func TestIntegrantes_BuscarPorUsuario(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, proyectoB := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)
	_, _ = repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Ana", "ana@mail.com", member.RolProductBuilder))
	_, _ = repo.VincularPorEmail(ctx, "ana@mail.com", usuarioAna)

	i, err := repo.BuscarPorUsuario(ctx, proyectoA, usuarioAna)
	if err != nil || i.Nombre != "Ana" || i.AuthUserID != usuarioAna {
		t.Errorf("se esperaba a Ana vinculada y se obtuvo %+v, %v", i, err)
	}
	for nombre, c := range map[string]struct {
		proyecto int64
		usuario  string
	}{"otro proyecto": {proyectoB, usuarioAna}, "otro usuario": {proyectoA, usuarioOtro}, "usuario vacío": {proyectoA, ""}} {
		if _, err := repo.BuscarPorUsuario(ctx, c.proyecto, c.usuario); !errors.Is(err, member.ErrNoEncontrado) {
			t.Errorf("%s: se esperaba member.ErrNoEncontrado y se obtuvo %v", nombre, err)
		}
	}
}

// El vínculo se lee con el resto de los datos, y dar de baja con una copia leída antes de vincular no lo borra
func TestIntegrantes_ElVinculoSeLeeYActualizarNoLoBorra(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)
	leido, _ := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Ana", "ana@mail.com", member.RolProductBuilder))
	_, _ = repo.VincularPorEmail(ctx, "ana@mail.com", usuarioAna)

	leido.Activo = false
	if err := repo.Actualizar(ctx, leido); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	lista, err := repo.ListarPorProyecto(ctx, proyectoA)
	if err != nil || len(lista) != 1 || lista[0].AuthUserID != usuarioAna || lista[0].Activo {
		t.Errorf("se esperaba a Ana de baja y vinculada y se obtuvo %+v, %v", lista, err)
	}
}
