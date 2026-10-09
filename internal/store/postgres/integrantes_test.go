package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

func integranteDeEjemplo(proyectoID int64, nombre, email string, rol member.Rol) member.Integrante {
	return member.Integrante{ProyectoID: proyectoID, Nombre: nombre, Email: email, Rol: rol, Activo: true}
}

func TestIntegrantes_GuardarAsignaIDsYListarPorProyecto(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, proyectoB := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)

	a, err := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Ana", "ana@mail.com", member.RolProductBuilder))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if _, err := repo.Guardar(ctx, integranteDeEjemplo(proyectoB, "De otro proyecto", "otro@mail.com", member.RolProductBuilder)); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	c, err := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Bea", "bea@mail.com", member.RolProductBuilder))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if a.ID == 0 || c.ID != a.ID+2 {
		t.Errorf("IDs: se esperaban correlativos (el del medio es del otro proyecto) y se obtuvo %d y %d", a.ID, c.ID)
	}
	lista, err := repo.ListarPorProyecto(ctx, proyectoA)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].Nombre != "Ana" || lista[1].Nombre != "Bea" {
		t.Errorf("se esperaban solo los integrantes del proyecto A, en orden, y se obtuvo %+v", lista)
	}
}

func TestIntegrantes_GuardaYDevuelveTodosLosCampos(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)
	original := integranteDeEjemplo(proyectoA, "Juan Pablo", "juanpablo@mail.com", member.RolAgileEnabler)

	guardado, err := repo.Guardar(ctx, original)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	leido, err := repo.BuscarPorID(ctx, proyectoA, guardado.ID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	original.ID = guardado.ID
	if leido != original {
		t.Errorf("se guardó %+v y se leyó %+v", original, leido)
	}
}

func TestIntegrantes_ProyectoSinIntegrantes_DevuelveListaVacia(t *testing.T) {
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	lista, err := postgres.NuevoIntegrantes(pool).ListarPorProyecto(context.Background(), proyectoA)
	if err != nil || len(lista) != 0 {
		t.Errorf("se esperaba una lista vacía sin error y se obtuvo %+v, %v", lista, err)
	}
}

// Un integrante se busca dentro de su proyecto: con el ID de otro proyecto no se encuentra.
func TestIntegrantes_BuscarPorID(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, proyectoB := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)
	guardado, _ := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Ana", "ana@mail.com", member.RolProductBuilder))

	i, err := repo.BuscarPorID(ctx, proyectoA, guardado.ID)
	if err != nil || i.Nombre != "Ana" {
		t.Errorf("se esperaba a Ana sin error y se obtuvo %+v, %v", i, err)
	}
	casos := map[string][2]int64{"ID inexistente": {proyectoA, guardado.ID + 99}, "otro proyecto": {proyectoB, guardado.ID}}
	for nombre, ids := range casos {
		if _, err := repo.BuscarPorID(ctx, ids[0], ids[1]); !errors.Is(err, member.ErrNoEncontrado) {
			t.Errorf("%s: se esperaba member.ErrNoEncontrado y se obtuvo %v", nombre, err)
		}
	}
}

// HU-03 RN7 · un integrante nunca se borra: la baja lo deja inactivo y sigue en la lista.
func TestIntegrantes_ActualizarGuardaLaBaja(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)
	guardado, _ := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Ana", "ana@mail.com", member.RolProductBuilder))

	guardado.Activo = false
	if err := repo.Actualizar(ctx, guardado); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	lista, err := repo.ListarPorProyecto(ctx, proyectoA)
	if err != nil || len(lista) != 1 || lista[0].Activo {
		t.Errorf("se esperaba un integrante inactivo que sigue en la lista y se obtuvo %+v, %v", lista, err)
	}
}

func TestIntegrantes_ActualizarUnIntegranteInexistenteOAjenoDevuelveErrNoEncontrado(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, proyectoB := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)
	guardado, _ := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Ana", "ana@mail.com", member.RolProductBuilder))

	inexistente := guardado
	inexistente.ID += 99
	ajeno := guardado
	ajeno.ProyectoID = proyectoB
	for nombre, i := range map[string]member.Integrante{"ID inexistente": inexistente, "de otro proyecto": ajeno} {
		if err := repo.Actualizar(ctx, i); !errors.Is(err, member.ErrNoEncontrado) {
			t.Errorf("%s: se esperaba member.ErrNoEncontrado y se obtuvo %v", nombre, err)
		}
	}
}

// El caso de uso ya controla estas reglas; la base es la última defensa si dos pedidos llegan a la vez.
func TestIntegrantes_LaBaseRechazaEmailRepetidoYSegundoAgileEnablerActivo(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, proyectoB := dosProyectos(t, pool)
	repo := postgres.NuevoIntegrantes(pool)
	ae, err := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "JP", "jp@mail.com", member.RolAgileEnabler))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if _, err := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Otro", "jp@mail.com", member.RolProductBuilder)); err == nil {
		t.Error("el email repetido en el proyecto debería rechazarse")
	}
	if _, err := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Otro AE", "ae2@mail.com", member.RolAgileEnabler)); err == nil {
		t.Error("un segundo Agile Enabler activo debería rechazarse")
	}
	if _, err := repo.Guardar(ctx, integranteDeEjemplo(proyectoB, "JP", "jp@mail.com", member.RolAgileEnabler)); err != nil {
		t.Errorf("el mismo email y rol en otro proyecto sí debería poder guardarse: %v", err)
	}

	// con el Agile Enabler dado de baja se puede registrar otro, pero su email sigue ocupado
	ae.Activo = false
	if err := repo.Actualizar(ctx, ae); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if _, err := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Nuevo AE", "ae2@mail.com", member.RolAgileEnabler)); err != nil {
		t.Errorf("con el anterior de baja se debería poder registrar otro Agile Enabler: %v", err)
	}
	if _, err := repo.Guardar(ctx, integranteDeEjemplo(proyectoA, "Repetido", "jp@mail.com", member.RolProductBuilder)); err == nil {
		t.Error("el email de un integrante dado de baja sigue ocupado")
	}
}

func TestIntegrantes_NoSeGuardaUnIntegranteDeUnProyectoQueNoExiste(t *testing.T) {
	repo := postgres.NuevoIntegrantes(baseMigrada(t))
	if _, err := repo.Guardar(context.Background(), integranteDeEjemplo(999, "Huérfano", "h@mail.com", member.RolProductBuilder)); err == nil {
		t.Error("se esperaba un error por el proyecto inexistente y se guardó")
	}
}
