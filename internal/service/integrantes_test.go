package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// escenarioIntegrantes arma los repositorios en memoria con el proyecto de ejemplo ya creado.
func escenarioIntegrantes(t *testing.T) (*service.Integrantes, *memory.Integrantes, int64) {
	t.Helper()
	proyectos := memory.NuevoProyectos()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(context.Background(), datosValidos())
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: %v", err)
	}
	integrantes := memory.NuevoIntegrantes()
	return service.NuevoIntegrantes(proyectos, integrantes), integrantes, p.ID
}

func datosDeIntegrante(email string, rol member.Rol) member.Datos {
	return member.Datos{Nombre: "Alguien", Email: email, Rol: rol}
}

// CA-03.1 y CA-03.2 · el integrante queda guardado activo, con ID, en su proyecto
func TestRegistrarIntegrante_DatosValidos_LoGuarda(t *testing.T) {
	s, repo, proyectoID := escenarioIntegrantes(t)

	i, err := s.Registrar(context.Background(), proyectoID, datosDeIntegrante("JP@mail.com", member.RolAgileEnabler))

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if i.ID == 0 || i.ProyectoID != proyectoID || !i.Activo || i.Email != "jp@mail.com" {
		t.Errorf("se esperaba el integrante guardado, activo y con el email normalizado, y se obtuvo %+v", i)
	}
	if lista, _ := repo.ListarPorProyecto(context.Background(), proyectoID); len(lista) != 1 {
		t.Errorf("se esperaba 1 integrante guardado y hay %d", len(lista))
	}
}

// El caso de uso le pasa al dominio los integrantes que ya tiene el proyecto (RN4 y RN6).
func TestRegistrarIntegrante_ConLosQueYaEstan_ControlaEmailYAgileEnabler(t *testing.T) {
	s, repo, proyectoID := escenarioIntegrantes(t)
	_, _ = s.Registrar(context.Background(), proyectoID, datosDeIntegrante("jp@mail.com", member.RolAgileEnabler))

	_, err := s.Registrar(context.Background(), proyectoID, datosDeIntegrante("JP@mail.com", member.RolAgileEnabler))

	if got := errorDeCampo(t, err, member.CampoEmail); !errors.Is(got, member.ErrEmailRepetido) {
		t.Errorf("email: se esperaba %v y se obtuvo %v", member.ErrEmailRepetido, got)
	}
	if got := errorDeCampo(t, err, member.CampoRol); !errors.Is(got, member.ErrAgileEnablerRepetido) {
		t.Errorf("rol: se esperaba %v y se obtuvo %v", member.ErrAgileEnablerRepetido, got)
	}
	if lista, _ := repo.ListarPorProyecto(context.Background(), proyectoID); len(lista) != 1 {
		t.Errorf("se esperaba que siga habiendo 1 integrante y hay %d", len(lista))
	}
}

func TestRegistrarIntegrante_ProyectoInexistente_DevuelveNoEncontrado(t *testing.T) {
	s, _, _ := escenarioIntegrantes(t)

	_, err := s.Registrar(context.Background(), 99, datosDeIntegrante("jp@mail.com", member.RolAgileEnabler))

	if !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("se esperaba project.ErrNoEncontrado y se obtuvo: %v", err)
	}
}

func TestListarIntegrantes_DevuelveElProyectoYSusIntegrantes(t *testing.T) {
	s, _, proyectoID := escenarioIntegrantes(t)
	_, _ = s.Registrar(context.Background(), proyectoID, datosDeIntegrante("jp@mail.com", member.RolAgileEnabler))

	p, lista, err := s.Listar(context.Background(), proyectoID)

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if p.ID != proyectoID || len(lista) != 1 || lista[0].Email != "jp@mail.com" {
		t.Errorf("se esperaba el proyecto %d con un integrante y se obtuvo %+v, %+v", proyectoID, p, lista)
	}
}

// CA-03.3 · la baja deja al integrante inactivo y guardado; la segunda vez da error
func TestDarDeBaja_QuedaInactivoYNoSeBorra(t *testing.T) {
	s, repo, proyectoID := escenarioIntegrantes(t)
	i, _ := s.Registrar(context.Background(), proyectoID, datosDeIntegrante("mp@mail.com", member.RolProductBuilder))

	baja, err := s.DarDeBaja(context.Background(), proyectoID, i.ID)

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	guardado, err := repo.BuscarPorID(context.Background(), proyectoID, i.ID)
	if err != nil || guardado.Activo || baja.Activo {
		t.Errorf("se esperaba que siga guardado e inactivo y se obtuvo %+v, %v", guardado, err)
	}
	if _, err := s.DarDeBaja(context.Background(), proyectoID, i.ID); !errors.Is(err, member.ErrYaDadoDeBaja) {
		t.Errorf("segunda baja: se esperaba %v y se obtuvo %v", member.ErrYaDadoDeBaja, err)
	}
}

func TestDarDeBaja_IntegranteInexistente_DevuelveNoEncontrado(t *testing.T) {
	s, _, proyectoID := escenarioIntegrantes(t)

	_, err := s.DarDeBaja(context.Background(), proyectoID, 99)

	if !errors.Is(err, member.ErrNoEncontrado) {
		t.Errorf("se esperaba member.ErrNoEncontrado y se obtuvo: %v", err)
	}
}

// integrantesQueFallan simula una base de datos caída.
type integrantesQueFallan struct{ errListar, errGuardar, errBuscar, errActualizar error }

func (r integrantesQueFallan) ListarPorProyecto(context.Context, int64) ([]member.Integrante, error) {
	return nil, r.errListar
}
func (r integrantesQueFallan) Guardar(_ context.Context, i member.Integrante) (member.Integrante, error) {
	return i, r.errGuardar
}
func (r integrantesQueFallan) BuscarPorID(_ context.Context, proyectoID, id int64) (member.Integrante, error) {
	return member.Integrante{ID: id, ProyectoID: proyectoID, Activo: true}, r.errBuscar
}
func (r integrantesQueFallan) Actualizar(context.Context, member.Integrante) error {
	return r.errActualizar
}

// Un error del repositorio no es de validación: se devuelve para que la pantalla muestre un error general.
func TestIntegrantes_ErrorDelRepositorio_SeDevuelve(t *testing.T) {
	falla := errors.New("base caída")
	ctx := context.Background()
	d := datosDeIntegrante("jp@mail.com", member.RolAgileEnabler)
	casos := map[string]func() error{
		"registrar, al buscar el proyecto": func() error {
			_, err := service.NuevoIntegrantes(proyectosQueFallan{err: falla}, integrantesQueFallan{}).Registrar(ctx, 1, d)
			return err
		},
		"registrar, al listar": func() error {
			_, err := service.NuevoIntegrantes(proyectoDeEjemplo{}, integrantesQueFallan{errListar: falla}).Registrar(ctx, 1, d)
			return err
		},
		"registrar, al guardar": func() error {
			_, err := service.NuevoIntegrantes(proyectoDeEjemplo{}, integrantesQueFallan{errGuardar: falla}).Registrar(ctx, 1, d)
			return err
		},
		"listar": func() error {
			_, _, err := service.NuevoIntegrantes(proyectoDeEjemplo{}, integrantesQueFallan{errListar: falla}).Listar(ctx, 1)
			return err
		},
		"baja, al buscar": func() error {
			_, err := service.NuevoIntegrantes(proyectoDeEjemplo{}, integrantesQueFallan{errBuscar: falla}).DarDeBaja(ctx, 1, 1)
			return err
		},
		"baja, al actualizar": func() error {
			_, err := service.NuevoIntegrantes(proyectoDeEjemplo{}, integrantesQueFallan{errActualizar: falla}).DarDeBaja(ctx, 1, 1)
			return err
		},
	}
	for nombre, caso := range casos {
		if err := caso(); !errors.Is(err, falla) {
			t.Errorf("%s: se esperaba el error del repositorio y se obtuvo: %v", nombre, err)
		}
	}
}
