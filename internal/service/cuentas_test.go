package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/auth"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/account"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// Los usuarios son los UUID que Supabase Auth le da a cada cuenta.
const (
	idAna  = "11111111-1111-4111-8111-111111111111"
	idBeto = "22222222-2222-4222-8222-222222222222"
)

// loginFalso hace de Supabase Auth: guarda las cuentas en memoria y cuenta las veces que lo consultaron.
type loginFalso struct {
	clavePorEmail   map[string]string
	idPorEmail      map[string]string
	pideConfirmar   bool
	sinConfirmar    map[string]bool
	falla           error // si no es nil, todo lo que se le pide falla con este error
	registros       []account.Credenciales
	ingresosPedidos int
}

func nuevoLoginFalso() *loginFalso {
	return &loginFalso{clavePorEmail: map[string]string{}, idPorEmail: map[string]string{}, sinConfirmar: map[string]bool{}}
}

// conCuenta crea una cuenta que ya existe en el servicio.
func (l *loginFalso) conCuenta(email, clave, id string) *loginFalso {
	l.clavePorEmail[email], l.idPorEmail[email] = clave, id
	return l
}

func (l *loginFalso) Registrar(_ context.Context, c account.Credenciales) (auth.ResultadoRegistro, error) {
	l.registros = append(l.registros, c)
	if l.falla != nil {
		return auth.ResultadoRegistro{}, l.falla
	}
	if _, existe := l.clavePorEmail[c.Email]; existe {
		return auth.ResultadoRegistro{}, auth.ErrEmailYaRegistrado
	}
	l.conCuenta(c.Email, c.Clave, idAna)
	return auth.ResultadoRegistro{RequiereConfirmacion: l.pideConfirmar}, nil
}

func (l *loginFalso) Ingresar(_ context.Context, c account.Credenciales) (auth.Sesion, error) {
	l.ingresosPedidos++
	if l.falla != nil {
		return auth.Sesion{}, l.falla
	}
	if clave, existe := l.clavePorEmail[c.Email]; !existe || clave != c.Clave {
		return auth.Sesion{}, auth.ErrCredencialesIncorrectas
	}
	if l.sinConfirmar[c.Email] {
		return auth.Sesion{}, auth.ErrEmailSinConfirmar
	}
	return auth.Sesion{
		Token:   "token-de-" + c.Email,
		Vence:   ahoraFijo.Add(8 * time.Hour),
		Usuario: auth.Identidad{ID: l.idPorEmail[c.Email], Email: c.Email},
	}, nil
}

// vinculosQueFallan es un repositorio de vínculos que no funciona: la base no responde.
type vinculosQueFallan struct{}

func (vinculosQueFallan) VincularPorEmail(context.Context, string, string) (int, error) {
	return 0, errors.New("la base no responde")
}

func (vinculosQueFallan) BuscarPorUsuario(context.Context, int64, string) (member.Integrante, error) {
	return member.Integrante{}, errors.New("la base no responde")
}

func datosDeRegistro(email, clave string) account.DatosRegistro {
	return account.DatosRegistro{Email: email, Clave: clave, Repeticion: clave}
}

// integranteEn agrega un integrante con ese email al proyecto de un repositorio en memoria.
func integranteEn(t *testing.T, repo *memory.Integrantes, proyectoID int64, nombre, email string) member.Integrante {
	t.Helper()
	i, err := repo.Guardar(context.Background(), member.Integrante{ProyectoID: proyectoID, Nombre: nombre, Email: email, Rol: member.RolProductBuilder, Activo: true})
	if err != nil {
		t.Fatalf("guardar el integrante %q: %v", nombre, err)
	}
	return i
}

// CA-05.1 · RN5: se valida, se normaliza el email y se le pide a Supabase crear la cuenta
func TestRegistrarCuenta_DatosValidos_CreaLaCuenta(t *testing.T) {
	login := nuevoLoginFalso()
	s := service.NuevoCuentas(login, memory.NuevoIntegrantes())

	res, err := s.Registrar(context.Background(), datosDeRegistro(" Ana@Mail.com ", "clave-segura-1"))

	if err != nil || res.RequiereConfirmacion {
		t.Fatalf("se esperaba la cuenta creada sin confirmación y se obtuvo %+v, %v", res, err)
	}
	if len(login.registros) != 1 || login.registros[0] != (account.Credenciales{Email: "ana@mail.com", Clave: "clave-segura-1"}) {
		t.Errorf("se esperaba un pedido con el email normalizado y se obtuvo %+v", login.registros)
	}
}

// RN5 · si Supabase pide confirmar el email, el caso de uso lo informa
func TestRegistrarCuenta_ConConfirmacionDeEmail_LoInforma(t *testing.T) {
	login := nuevoLoginFalso()
	login.pideConfirmar = true

	res, err := service.NuevoCuentas(login, memory.NuevoIntegrantes()).Registrar(context.Background(), datosDeRegistro("ana@mail.com", "clave-segura-1"))

	if err != nil || !res.RequiereConfirmacion {
		t.Errorf("se esperaba que pida confirmar el email y se obtuvo %+v, %v", res, err)
	}
}

// RN5 · con datos inválidos se informan todos los errores y no se le pide nada a Supabase
func TestRegistrarCuenta_DatosInvalidos_NoLlamaAlServicio(t *testing.T) {
	login := nuevoLoginFalso()
	s := service.NuevoCuentas(login, memory.NuevoIntegrantes())

	_, err := s.Registrar(context.Background(), account.DatosRegistro{Email: "ana@mail", Clave: "corta", Repeticion: "otra"})

	for _, campo := range []string{account.CampoEmail, account.CampoClave, account.CampoRepeticion} {
		if errorDeCampo(t, err, campo) == nil {
			t.Errorf("se esperaba un error en el campo %q", campo)
		}
	}
	if len(login.registros) != 0 {
		t.Errorf("no se tendría que haber consultado a Supabase y hubo %d pedidos", len(login.registros))
	}
}

// RN5 · los errores de Supabase que son del formulario se muestran en su campo
func TestRegistrarCuenta_ErroresDeSupabaseDelFormulario_VanEnSuCampo(t *testing.T) {
	casos := []struct {
		nombre   string
		falla    error
		campo    string
		esperado error
	}{
		{"email ya registrado", auth.ErrEmailYaRegistrado, account.CampoEmail, auth.ErrEmailYaRegistrado},
		{"contraseña débil", auth.ErrClaveDebil, account.CampoClave, auth.ErrClaveDebil},
		{"email rechazado", auth.ErrEmailRechazado, account.CampoEmail, member.ErrEmailInvalido},
	}
	for _, c := range casos {
		login := nuevoLoginFalso()
		login.falla = c.falla

		_, err := service.NuevoCuentas(login, memory.NuevoIntegrantes()).Registrar(context.Background(), datosDeRegistro("ana@mail.com", "clave-segura-1"))

		if got := errorDeCampo(t, err, c.campo); !errors.Is(got, c.esperado) {
			t.Errorf("%s: se esperaba %v en %q y se obtuvo %v", c.nombre, c.esperado, c.campo, got)
		}
	}
}

// RN12 · si el servicio no responde o limita los intentos, el error pasa tal cual para que la pantalla elija el mensaje
func TestRegistrarEIngresar_ServicioCaidoOLimitado_DevuelvenElError(t *testing.T) {
	for _, falla := range []error{auth.ErrServicioNoDisponible, auth.ErrDemasiadosIntentos} {
		login := nuevoLoginFalso()
		login.falla = falla
		s := service.NuevoCuentas(login, memory.NuevoIntegrantes())

		if _, err := s.Registrar(context.Background(), datosDeRegistro("ana@mail.com", "clave-segura-1")); !errors.Is(err, falla) {
			t.Errorf("al registrar se esperaba %v y se obtuvo %v", falla, err)
		}
		if _, err := s.Ingresar(context.Background(), "ana@mail.com", "clave-segura-1"); !errors.Is(err, falla) {
			t.Errorf("al ingresar se esperaba %v y se obtuvo %v", falla, err)
		}
	}
}

// CA-05.1 y CA-05.3 · RN6 y RN7: al ingresar se devuelve la sesión y el usuario queda vinculado con sus integrantes de todos
// los proyectos, y con ningún otro
func TestIngresar_CredencialesCorrectas_DevuelveLaSesionYVinculaALosIntegrantes(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	ana1 := integranteEn(t, repo, 1, "Ana", "ana@mail.com")
	ana2 := integranteEn(t, repo, 2, "Ana", "ana@mail.com")
	beto := integranteEn(t, repo, 1, "Beto", "beto@mail.com")
	s := service.NuevoCuentas(nuevoLoginFalso().conCuenta("ana@mail.com", "clave-segura-1", idAna), repo)

	sesion, err := s.Ingresar(ctx, "Ana@Mail.com", "clave-segura-1")

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if sesion.Token != "token-de-ana@mail.com" || sesion.Usuario.ID != idAna {
		t.Errorf("se esperaba la sesión de Ana y se obtuvo %+v", sesion)
	}
	for _, i := range []member.Integrante{ana1, ana2} {
		if got, err := repo.BuscarPorID(ctx, i.ProyectoID, i.ID); err != nil || got.AuthUserID != idAna {
			t.Errorf("Ana en el proyecto %d tendría que estar vinculada y se obtuvo %+v, %v", i.ProyectoID, got, err)
		}
	}
	if got, _ := repo.BuscarPorID(ctx, beto.ProyectoID, beto.ID); got.AuthUserID != "" {
		t.Errorf("Beto no tendría que estar vinculado y tiene %q", got.AuthUserID)
	}
}

// RN6 · con credenciales incorrectas no hay sesión ni vínculo; con el email sin confirmar el error es otro
func TestIngresar_CredencialesIncorrectasOSinConfirmar_NoVincula(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	ana := integranteEn(t, repo, 1, "Ana", "ana@mail.com")
	login := nuevoLoginFalso().conCuenta("ana@mail.com", "clave-segura-1", idAna)
	login.sinConfirmar["ana@mail.com"] = true
	s := service.NuevoCuentas(login, repo)

	if _, err := s.Ingresar(ctx, "ana@mail.com", "otra-clave-9"); !errors.Is(err, auth.ErrCredencialesIncorrectas) {
		t.Errorf("con otra contraseña se esperaba auth.ErrCredencialesIncorrectas y se obtuvo %v", err)
	}
	if _, err := s.Ingresar(ctx, "nadie@mail.com", "clave-segura-1"); !errors.Is(err, auth.ErrCredencialesIncorrectas) {
		t.Errorf("con otro email se esperaba auth.ErrCredencialesIncorrectas y se obtuvo %v", err)
	}
	if _, err := s.Ingresar(ctx, "ana@mail.com", "clave-segura-1"); !errors.Is(err, auth.ErrEmailSinConfirmar) {
		t.Errorf("sin confirmar se esperaba auth.ErrEmailSinConfirmar y se obtuvo %v", err)
	}
	if got, _ := repo.BuscarPorID(ctx, ana.ProyectoID, ana.ID); got.AuthUserID != "" {
		t.Errorf("sin ingreso correcto Ana no tendría que quedar vinculada y tiene %q", got.AuthUserID)
	}
}

// RN6 · sin email o sin contraseña no hay nada que consultar: son credenciales incorrectas y no se llama a Supabase
func TestIngresar_CredencialesIncompletas_NoLlamaAlServicio(t *testing.T) {
	login := nuevoLoginFalso()
	s := service.NuevoCuentas(login, memory.NuevoIntegrantes())

	for _, c := range [][2]string{{"", "clave"}, {"ana@mail.com", ""}, {"  ", "  "}} {
		if _, err := s.Ingresar(context.Background(), c[0], c[1]); !errors.Is(err, auth.ErrCredencialesIncorrectas) {
			t.Errorf("con %q y %q se esperaba auth.ErrCredencialesIncorrectas y se obtuvo %v", c[0], c[1], err)
		}
	}
	if login.ingresosPedidos != 0 {
		t.Errorf("no se tendría que haber consultado a Supabase y hubo %d pedidos", login.ingresosPedidos)
	}
}

// RN7 · si no se puede vincular (la base no responde) el ingreso igual funciona: Quien lo vuelve a intentar (RN8)
func TestIngresar_SiFallaElVinculo_IgualIngresa(t *testing.T) {
	s := service.NuevoCuentas(nuevoLoginFalso().conCuenta("ana@mail.com", "clave-segura-1", idAna), vinculosQueFallan{})

	sesion, err := s.Ingresar(context.Background(), "ana@mail.com", "clave-segura-1")

	if err != nil || sesion.Token == "" {
		t.Errorf("se esperaba la sesión sin error y se obtuvo %+v, %v", sesion, err)
	}
}

// CA-05.3 · RN8: el integrante del proyecto vinculado con el usuario
func TestQuien_UsuarioVinculado_DevuelveSuIntegrante(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	ana := integranteEn(t, repo, 1, "Ana", "ana@mail.com")
	integranteEn(t, repo, 1, "Beto", "beto@mail.com")
	_, _ = repo.VincularPorEmail(ctx, "ana@mail.com", idAna)
	s := service.NuevoCuentas(nuevoLoginFalso(), repo)

	i, err := s.Quien(ctx, auth.Identidad{ID: idAna, Email: "ana@mail.com"}, 1)

	if err != nil || i.ID != ana.ID || i.Nombre != "Ana" {
		t.Errorf("se esperaba a Ana y se obtuvo %+v, %v", i, err)
	}
}

// RN8 · si lo registraron en el proyecto después de que ingresó, Quien lo vincula en ese momento
func TestQuien_IntegranteRegistradoDespuesDelIngreso_LoVincula(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	s := service.NuevoCuentas(nuevoLoginFalso(), repo)
	usuario := auth.Identidad{ID: idAna, Email: "ana@mail.com"}
	if _, err := s.Quien(ctx, usuario, 1); !errors.Is(err, member.ErrNoEsIntegrante) {
		t.Fatalf("todavía no es integrante y se obtuvo %v", err)
	}
	ana := integranteEn(t, repo, 1, "Ana", "ana@mail.com")

	i, err := s.Quien(ctx, usuario, 1)

	if err != nil || i.ID != ana.ID || i.AuthUserID != idAna {
		t.Errorf("se esperaba a Ana vinculada y se obtuvo %+v, %v", i, err)
	}
	if guardada, _ := repo.BuscarPorUsuario(ctx, 1, idAna); guardada.ID != ana.ID {
		t.Errorf("el vínculo tendría que quedar guardado y se obtuvo %+v", guardada)
	}
}

// RN8 · quien no es integrante del proyecto, o ya está vinculado con otro usuario, no es integrante; sin email tampoco
func TestQuien_NoEsIntegranteDelProyecto(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	integranteEn(t, repo, 1, "Beto", "beto@mail.com")
	integranteEn(t, repo, 2, "Ana", "ana@mail.com")
	integranteEn(t, repo, 3, "Ana de otra cuenta", "otra-ana@mail.com")
	_, _ = repo.VincularPorEmail(ctx, "otra-ana@mail.com", idBeto)
	s := service.NuevoCuentas(nuevoLoginFalso(), repo)

	casos := map[string]struct {
		usuario  auth.Identidad
		proyecto int64
	}{
		"otro email":                       {auth.Identidad{ID: idAna, Email: "ana@mail.com"}, 1},
		"integrante solo de otro proyecto": {auth.Identidad{ID: idBeto, Email: "beto@mail.com"}, 2},
		"vinculado con otro usuario":       {auth.Identidad{ID: idAna, Email: "otra-ana@mail.com"}, 3},
		"sin email":                        {auth.Identidad{ID: idAna}, 1},
		"proyecto que no existe":           {auth.Identidad{ID: idAna, Email: "ana@mail.com"}, 99},
	}
	for nombre, c := range casos {
		if _, err := s.Quien(ctx, c.usuario, c.proyecto); !errors.Is(err, member.ErrNoEsIntegrante) {
			t.Errorf("%s: se esperaba member.ErrNoEsIntegrante y se obtuvo %v", nombre, err)
		}
	}
}

// RN8 · un integrante dado de baja viene con Activo en false: quien lo use decide qué hacer
func TestQuien_IntegranteDadoDeBaja_LoDevuelveInactivo(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	ana := integranteEn(t, repo, 1, "Ana", "ana@mail.com")
	ana.Activo = false
	_ = repo.Actualizar(ctx, ana)
	s := service.NuevoCuentas(nuevoLoginFalso(), repo)

	i, err := s.Quien(ctx, auth.Identidad{ID: idAna, Email: "ana@mail.com"}, 1)

	if err != nil || i.ID != ana.ID || i.Activo {
		t.Errorf("se esperaba a Ana de baja y se obtuvo %+v, %v", i, err)
	}
}

// Un error de la base no es lo mismo que no ser integrante
func TestQuien_ErrorDeLaBase_NoEsNoSerIntegrante(t *testing.T) {
	s := service.NuevoCuentas(nuevoLoginFalso(), vinculosQueFallan{})

	_, err := s.Quien(context.Background(), auth.Identidad{ID: idAna, Email: "ana@mail.com"}, 1)

	if err == nil || errors.Is(err, member.ErrNoEsIntegrante) {
		t.Errorf("se esperaba el error de la base y se obtuvo %v", err)
	}
}
