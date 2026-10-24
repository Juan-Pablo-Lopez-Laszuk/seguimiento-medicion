package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/auth"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/account"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// RepositorioVinculos es lo que Cuentas necesita de los integrantes para saber quién es cada usuario (TEC-05). Lo implementan
// internal/store/memory e internal/store/postgres.
type RepositorioVinculos interface {
	// VincularPorEmail vincula al usuario con los integrantes de cualquier proyecto que tienen ese email y todavía no tienen
	// un usuario, y devuelve cuántos vinculó.
	VincularPorEmail(ctx context.Context, email, usuarioID string) (int, error)
	// BuscarPorUsuario devuelve el integrante del proyecto vinculado con el usuario, o member.ErrNoEncontrado.
	BuscarPorUsuario(ctx context.Context, proyectoID int64, usuarioID string) (member.Integrante, error)
}

// Cuentas agrupa los casos de uso de las cuentas de usuario: registrarse, ingresar y saber a qué integrante de un proyecto
// corresponde el usuario (TEC-05).
type Cuentas struct {
	servicio auth.ServicioDeCuentas
	vinculos RepositorioVinculos
}

// NuevoCuentas arma el caso de uso con el servicio que crea las cuentas (Supabase Auth) y el repositorio de integrantes.
func NuevoCuentas(servicio auth.ServicioDeCuentas, vinculos RepositorioVinculos) *Cuentas {
	return &Cuentas{servicio: servicio, vinculos: vinculos}
}

// Registrar valida los datos y crea la cuenta en el servicio (RN5). Los datos inválidos, incluidos los que rechaza el
// servicio (el email ya tiene cuenta, la contraseña es débil), se devuelven juntos en domain.ErroresValidacion. Si el
// servicio no responde o limita los intentos, el error es auth.ErrServicioNoDisponible o auth.ErrDemasiadosIntentos.
func (s *Cuentas) Registrar(ctx context.Context, d account.DatosRegistro) (auth.ResultadoRegistro, error) {
	credenciales, err := account.NuevoRegistro(d)
	if err != nil {
		return auth.ResultadoRegistro{}, err
	}
	resultado, err := s.servicio.Registrar(ctx, credenciales)
	switch {
	case err == nil:
		return resultado, nil
	case errors.Is(err, auth.ErrEmailYaRegistrado):
		return auth.ResultadoRegistro{}, domain.ErroresValidacion{account.CampoEmail: auth.ErrEmailYaRegistrado}
	case errors.Is(err, auth.ErrEmailRechazado):
		return auth.ResultadoRegistro{}, domain.ErroresValidacion{account.CampoEmail: member.ErrEmailInvalido}
	case errors.Is(err, auth.ErrClaveDebil):
		return auth.ResultadoRegistro{}, domain.ErroresValidacion{account.CampoClave: auth.ErrClaveDebil}
	}
	return auth.ResultadoRegistro{}, fmt.Errorf("registrar la cuenta: %w", err)
}

// Ingresar abre la sesión con el email y la contraseña y vincula al usuario con los integrantes que tienen su email, de
// todos los proyectos (RN6 y RN7). Con credenciales incorrectas, o incompletas, el error es auth.ErrCredencialesIncorrectas
// sin decir cuál de los dos datos falló. Si no se puede vincular, el ingreso igual funciona: Quien lo vuelve a intentar.
func (s *Cuentas) Ingresar(ctx context.Context, email, clave string) (auth.Sesion, error) {
	credenciales, err := account.Ingreso(email, clave)
	if err != nil {
		return auth.Sesion{}, auth.ErrCredencialesIncorrectas
	}
	sesion, err := s.servicio.Ingresar(ctx, credenciales)
	if err != nil {
		return auth.Sesion{}, fmt.Errorf("ingresar: %w", err)
	}
	emailDeLaCuenta := sesion.Usuario.Email
	if emailDeLaCuenta == "" {
		emailDeLaCuenta = credenciales.Email
	}
	if _, err := s.vinculos.VincularPorEmail(ctx, emailDeLaCuenta, sesion.Usuario.ID); err != nil {
		log.Printf("cuentas: no se pudo vincular al usuario %s con sus integrantes: %v", sesion.Usuario.ID, err)
	}
	return sesion, nil
}

// Quien devuelve el integrante del proyecto que corresponde al usuario (RN8). Si todavía no está vinculado pero hay un
// integrante con su email —lo registraron después de que ingresó— lo vincula en ese momento. Si el usuario no es integrante
// del proyecto, el error es member.ErrNoEsIntegrante. Un integrante dado de baja se devuelve con Activo en false.
func (s *Cuentas) Quien(ctx context.Context, usuario auth.Identidad, proyectoID int64) (member.Integrante, error) {
	i, err := s.vinculos.BuscarPorUsuario(ctx, proyectoID, usuario.ID)
	if err == nil {
		return i, nil
	}
	if !errors.Is(err, member.ErrNoEncontrado) {
		return member.Integrante{}, fmt.Errorf("buscar el integrante del usuario en el proyecto %d: %w", proyectoID, err)
	}
	if usuario.Email == "" {
		return member.Integrante{}, member.ErrNoEsIntegrante
	}

	vinculados, err := s.vinculos.VincularPorEmail(ctx, usuario.Email, usuario.ID)
	if err != nil {
		return member.Integrante{}, fmt.Errorf("vincular al usuario con sus integrantes: %w", err)
	}
	if vinculados == 0 {
		return member.Integrante{}, member.ErrNoEsIntegrante
	}
	i, err = s.vinculos.BuscarPorUsuario(ctx, proyectoID, usuario.ID)
	switch {
	case errors.Is(err, member.ErrNoEncontrado):
		return member.Integrante{}, member.ErrNoEsIntegrante
	case err != nil:
		return member.Integrante{}, fmt.Errorf("buscar el integrante del usuario en el proyecto %d: %w", proyectoID, err)
	}
	return i, nil
}
