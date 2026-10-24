package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// Los usuarios son los UUID que Supabase Auth le da a cada cuenta (TEC-05).
const (
	usuarioAna  = "11111111-1111-4111-8111-111111111111"
	usuarioOtro = "22222222-2222-4222-8222-222222222222"
)

// CA-05.3 y RN7 · el usuario se vincula con todos los integrantes que tienen su email, de cualquier proyecto y estén o no
// dados de baja, y con ningún otro
func TestIntegrantes_VincularPorEmail_VinculaATodosLosQueTienenElEmail(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	_, _ = repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Ana", Email: "ana@mail.com", Activo: true})
	_, _ = repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Beto", Email: "beto@mail.com", Activo: true})
	_, _ = repo.Guardar(ctx, member.Integrante{ProyectoID: 2, Nombre: "Ana en Beta", Email: "ana@mail.com", Activo: true})
	_, _ = repo.Guardar(ctx, member.Integrante{ProyectoID: 3, Nombre: "Ana de baja", Email: "ana@mail.com", Activo: false})

	vinculados, err := repo.VincularPorEmail(ctx, "ana@mail.com", usuarioAna)

	if err != nil || vinculados != 3 {
		t.Fatalf("se esperaban 3 integrantes vinculados sin error y se obtuvo %d, %v", vinculados, err)
	}
	for _, proyecto := range []int64{1, 2, 3} {
		if i, err := repo.BuscarPorUsuario(ctx, proyecto, usuarioAna); err != nil || i.Nombre == "Beto" {
			t.Errorf("proyecto %d: se esperaba a Ana vinculada y se obtuvo %+v, %v", proyecto, i, err)
		}
	}
	beto, _ := repo.BuscarPorID(ctx, 1, 2)
	if beto.AuthUserID != "" {
		t.Errorf("Beto tiene otro email y no se tendría que vincular: %q", beto.AuthUserID)
	}
}

// El email se compara sin distinguir mayúsculas, como el resto de la aplicación
func TestIntegrantes_VincularPorEmail_NoDistingueMayusculas(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	_, _ = repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Ana", Email: "ana@mail.com", Activo: true})

	vinculados, err := repo.VincularPorEmail(ctx, "Ana@Mail.com", usuarioAna)

	if err != nil || vinculados != 1 {
		t.Errorf("se esperaba 1 integrante vinculado sin error y se obtuvo %d, %v", vinculados, err)
	}
}

// RN7 · un integrante ya vinculado con otro usuario no se pisa, y uno ya vinculado con este usuario no se cuenta de nuevo
func TestIntegrantes_VincularPorEmail_NoPisaAUnIntegranteYaVinculado(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	_, _ = repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Ana", Email: "ana@mail.com", Activo: true})
	_, _ = repo.Guardar(ctx, member.Integrante{ProyectoID: 2, Nombre: "Ana en Beta", Email: "ana@mail.com", Activo: true})
	if _, err := repo.VincularPorEmail(ctx, "ana@mail.com", usuarioOtro); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	vinculados, err := repo.VincularPorEmail(ctx, "ana@mail.com", usuarioAna)

	if err != nil || vinculados != 0 {
		t.Errorf("se esperaban 0 integrantes vinculados sin error y se obtuvo %d, %v", vinculados, err)
	}
	if i, err := repo.BuscarPorUsuario(ctx, 1, usuarioOtro); err != nil || i.Nombre != "Ana" {
		t.Errorf("Ana tenía que seguir vinculada con el otro usuario y se obtuvo %+v, %v", i, err)
	}
	if _, err := repo.BuscarPorUsuario(ctx, 1, usuarioAna); !errors.Is(err, member.ErrNoEncontrado) {
		t.Errorf("el usuario nuevo no tendría que estar vinculado y se obtuvo %v", err)
	}
	if vinculados, _ := repo.VincularPorEmail(ctx, "ana@mail.com", usuarioOtro); vinculados != 0 {
		t.Errorf("volver a vincular al mismo usuario no cambia nada y se contaron %d", vinculados)
	}
}

func TestIntegrantes_VincularPorEmail_SinCoincidenciasONoHayUsuario(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	_, _ = repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Ana", Email: "ana@mail.com", Activo: true})

	if vinculados, err := repo.VincularPorEmail(ctx, "nadie@mail.com", usuarioAna); err != nil || vinculados != 0 {
		t.Errorf("sin coincidencias se esperaban 0 vinculados sin error y se obtuvo %d, %v", vinculados, err)
	}
	if _, err := repo.VincularPorEmail(ctx, "ana@mail.com", ""); err == nil {
		t.Error("sin usuario se esperaba un error")
	}
}

// BuscarPorUsuario busca dentro del proyecto: con el de otro proyecto o con un usuario sin vínculo no se encuentra
func TestIntegrantes_BuscarPorUsuario(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	_, _ = repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Ana", Email: "ana@mail.com", Activo: true})
	_, _ = repo.VincularPorEmail(ctx, "ana@mail.com", usuarioAna)

	i, err := repo.BuscarPorUsuario(ctx, 1, usuarioAna)
	if err != nil || i.Nombre != "Ana" || i.AuthUserID != usuarioAna {
		t.Errorf("se esperaba a Ana vinculada y se obtuvo %+v, %v", i, err)
	}
	for nombre, c := range map[string]struct {
		proyecto int64
		usuario  string
	}{"otro proyecto": {2, usuarioAna}, "otro usuario": {1, usuarioOtro}, "usuario vacío": {1, ""}} {
		if _, err := repo.BuscarPorUsuario(ctx, c.proyecto, c.usuario); !errors.Is(err, member.ErrNoEncontrado) {
			t.Errorf("%s: se esperaba member.ErrNoEncontrado y se obtuvo %v", nombre, err)
		}
	}
}

// Dar de baja a un integrante con una copia leída antes de vincularlo no borra el vínculo
func TestIntegrantes_ActualizarNoBorraElVinculo(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	leido, _ := repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Ana", Email: "ana@mail.com", Activo: true})
	_, _ = repo.VincularPorEmail(ctx, "ana@mail.com", usuarioAna)

	leido.Activo = false
	if err := repo.Actualizar(ctx, leido); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if i, err := repo.BuscarPorUsuario(ctx, 1, usuarioAna); err != nil || i.Activo {
		t.Errorf("se esperaba a Ana vinculada y de baja y se obtuvo %+v, %v", i, err)
	}
}
