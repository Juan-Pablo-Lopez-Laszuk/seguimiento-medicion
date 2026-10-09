# SPEC-03 · Registrar integrantes

**Historia:** HU-03 (#41) · **Responsable:** López, Juan Pablo · **Sprint:** 1 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md` y la tabla Integrante de
> `docs/modelo-datos.md`.

## 1. Objetivo

Que el Agile Enabler pueda registrar a las personas del equipo de un proyecto, con su nombre, email y rol, y darlas
de baja cuando dejan el proyecto. Los integrantes son los que después registran esfuerzo (HU-24), votan en el
Planning Poker (HU-18) y son responsables de tareas (HU-25).

## 2. Entradas

**Registrar un integrante**

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Nombre | texto | Sí | Hasta 100 caracteres (después de quitar espacios al inicio y al final) |
| Email | texto | Sí | Formato de email válido, hasta 254 caracteres; no repetido dentro del proyecto |
| Rol | opción | Sí | Agile Enabler o Product Builder; un solo Agile Enabler activo por proyecto |

**Dar de baja un integrante:** el proyecto y el integrante (vienen en la URL).

## 3. Salidas esperadas

- Al registrar: el integrante queda guardado con id, proyecto, nombre, email, rol y **activo**. En la interfaz se
  vuelve a la lista de integrantes del proyecto con el mensaje **"Integrante registrado"**.
- Al dar de baja: el integrante queda **inactivo** (no se borra) y en la lista aparece como "Dado de baja". Se vuelve
  a la lista con el mensaje **"Integrante dado de baja"**.

## 4. Reglas de negocio

- RN1: El nombre se guarda **sin los espacios del inicio y del final**; si queda vacío, es obligatorio. El largo se
  cuenta en caracteres.
- RN2: El email se guarda **sin espacios alrededor y en minúsculas**, así "Ana@Mail.com" y "ana@mail.com" son el
  mismo email (CA-03.1). Esto también sirve para vincular al integrante con su usuario en el primer login (TEC-05,
  decisión 1 de `docs/modelo-datos.md`).
- RN3: El email tiene que tener **formato válido**: una sola dirección, sin nombre delante (`Ana <ana@mail.com>` no
  vale), con usuario, `@` y un dominio con al menos un punto (`ana@mail` no vale).
- RN4: El email **no se repite dentro del proyecto**, contando también a los dados de baja (la base tiene el índice
  único `(proyecto_id, email)`). El mismo email sí puede estar en otro proyecto (CA-03.1).
- RN5: El rol es **Agile Enabler** o **Product Builder** (CA-03.2). Cualquier otro valor es un error.
- RN6: Hay **un solo Agile Enabler activo** por proyecto (CA-03.2). Si el Agile Enabler se da de baja, se puede
  registrar otro.
- RN7: Un integrante **nunca se borra: se da de baja** (queda inactivo). Así se conserva su historial de esfuerzo
  (CA-03.3), sus votos y sus tareas. Como HU-24 todavía no existe, la regla se aplica a todos los integrantes, tengan
  esfuerzo registrado o no.
- RN8: Dar de baja a alguien que ya está dado de baja es un error ("el integrante ya está dado de baja"). En esta
  historia no se reactiva a nadie.
- RN9: Si hay varios datos inválidos, se informan **todos juntos**, cada uno con su campo (como en HU-01).

## 5. Restricciones

- Las reglas de un integrante (RN1 a RN3, RN5 y RN8) son **funciones puras** en `internal/domain/member`. Las que
  necesitan a los demás integrantes del proyecto (RN4 y RN6) también son puras: el dominio recibe la lista de
  integrantes que ya tiene el proyecto, como en HU-11.
- El caso de uso en `internal/service` busca el proyecto y sus integrantes en los repositorios, llama al dominio y
  guarda. **Para TEC-03:** hace falta un repositorio de integrantes con `ListarPorProyecto`, `Guardar` y
  `Actualizar`, y el índice único parcial del Agile Enabler pasa a ser `(proyecto_id) WHERE rol = 'AgileEnabler' AND
  activo`, para que uno dado de baja no impida registrar otro (RN6).
- `auth_user_id` (TEC-05) no se toca en esta historia.
- Hasta que exista la vista del proyecto (HU-04, Sprint 2), a los integrantes se llega desde la lista de proyectos.
- La baja se pide con un formulario `POST` (no con un enlace `GET`), así un buscador o una vista previa no dan de baja
  a nadie, y la pantalla pide confirmación antes de enviarlo.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Primer integrante del proyecto, Agile Enabler | Se registra activo |
| Nombre de solo espacios | Error: nombre obligatorio |
| Nombre de 100 caracteres / de 101 | Se registra / error de largo |
| Email " Ana@Mail.COM " | Se guarda "ana@mail.com" |
| Email "ana@mail", "ana.mail.com", "Ana <ana@mail.com>" o vacío | Error de formato (vacío: obligatorio) |
| Email repetido con otras mayúsculas | Error de email repetido |
| Email de un integrante dado de baja | Error de email repetido |
| Mismo email en otro proyecto | Se registra |
| Rol vacío o "Scrum Master" | Error de rol |
| Segundo Agile Enabler activo | Error: ya hay un Agile Enabler |
| Agile Enabler nuevo cuando el anterior está dado de baja | Se registra |
| Varios Product Builder | Se registran todos |
| Dar de baja a un integrante activo | Queda inactivo y sigue en la lista |
| Dar de baja a uno ya dado de baja | Error: ya está dado de baja |
| Proyecto o integrante que no existe (o de otro proyecto) | Página "no encontrado" (404) |

## 7. Condiciones de error

| Situación | Campo | Error (Go) | Mensaje al usuario |
|---|---|---|---|
| Nombre vacío o solo espacios | nombre | `ErrNombreObligatorio` | "el nombre es obligatorio" |
| Nombre de más de 100 caracteres | nombre | `ErrNombreLargo` | "el nombre no puede superar los 100 caracteres" |
| Email vacío | email | `ErrEmailObligatorio` | "el email es obligatorio" |
| Email con formato inválido o de más de 254 caracteres | email | `ErrEmailInvalido` | "el email no tiene un formato válido" |
| Email repetido en el proyecto | email | `ErrEmailRepetido` | "ya hay un integrante con ese email en el proyecto" |
| Rol vacío o desconocido | rol | `ErrRolInvalido` | "el rol tiene que ser Agile Enabler o Product Builder" |
| Ya hay un Agile Enabler activo | rol | `ErrAgileEnablerRepetido` | "el proyecto ya tiene un Agile Enabler activo" |
| Baja de un integrante ya dado de baja | — | `ErrYaDadoDeBaja` | "el integrante ya está dado de baja" (aviso en la lista) |
| El proyecto no existe | — | `project.ErrNoEncontrado` | Página 404 "No encontramos ese proyecto" |
| El integrante no existe o es de otro proyecto | — | `member.ErrNoEncontrado` | Página 404 "No encontramos ese integrante" |

## 8. Criterios de aceptación

- CA-03.1: El email es válido y no se repite dentro del proyecto.
- CA-03.2: El rol es Agile Enabler o Product Builder, con un solo Agile Enabler por proyecto.
- CA-03.3: Un integrante con esfuerzo registrado no se elimina: se da de baja.

## Trazabilidad

Issue: #41 · Feature: `features/hu-03-registrar-integrantes.feature` · Pasos BDD: `features/hu03_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Dominio | `internal/domain/member/member.go` (`Nuevo`, `DarDeBaja`, `Rol.Texto`) | `internal/domain/member/member_test.go` |
| Repositorio en memoria | `internal/store/memory/integrantes.go` | `internal/store/memory/integrantes_test.go` |
| Caso de uso | `internal/service/integrantes.go` (`Registrar`, `Listar`, `DarDeBaja`) | `internal/service/integrantes_test.go` |
| Pantalla | `internal/server/integrantes.go`, `web/templates/paginas/integrantes.html`, `integrante_nuevo.html` | `internal/server/integrantes_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-03.1 Email válido y no repetido | Registrar al Agile Enabler · Formato del email · No se repite el email dentro del proyecto · El mismo email en otro proyecto | `TestNuevo_DatosValidos_*`, `TestNuevo_FormatoDelEmail`, `TestNuevo_EmailRepetido`, `TestRegistrarIntegrante_ConLosQueYaEstan_*`, `TestRegistrarIntegrante_DatosInvalidos_*` |
| CA-03.2 Rol y un solo Agile Enabler | Registrar al Agile Enabler · Registrar varios Product Builder · No puede haber dos Agile Enabler activos · Nuevo Agile Enabler después de una baja · Se indica cada campo que falló | `TestNuevo_Rol`, `TestRol_Texto`, `TestNuevo_UnSoloAgileEnablerActivo`, `TestNuevo_InformaTodosLosCamposInvalidosJuntos`, `TestNuevo_Nombre`, `TestFormularioNuevoIntegrante_*` |
| CA-03.3 Se da de baja, no se elimina | Dar de baja a un integrante · No se da de baja dos veces | `TestDarDeBaja_IntegranteActivo_*`, `TestDarDeBaja_YaDadoDeBaja_*`, `TestIntegrantes_Actualizar`, `TestDarDeBaja_QuedaInactivoYNoSeBorra`, `TestDarDeBaja_VuelveALaListaConElIntegranteDadoDeBaja`, `TestDarDeBaja_DosVeces_MuestraElAviso` |
| Proyecto o integrante inexistente (sección 7) | — | `TestIntegrantes_BuscarPorID`, `TestRegistrarIntegrante_ProyectoInexistente_*`, `TestDarDeBaja_IntegranteInexistente_*`, `TestIntegrantes_NoEncontrado_Responde404` |
