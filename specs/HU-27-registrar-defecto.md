# SPEC-27 · Registrar defecto

**Historia:** HU-27 (#5) · **Responsable:** Bravo, Carolina · **Sprint:** 2 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`, la tabla Defecto de
> `docs/modelo-datos.md` y las métricas M8 y M9 de `docs/metricas.md`.

## 1. Objetivo

Que un integrante pueda registrar un defecto encontrado en el proyecto, con su severidad y el sprint en que se
detectó, para darle seguimiento (HU-28) y contar los defectos detectados por sprint (M8, HU-34).

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id (viene en la URL) | Sí | Tiene que existir |
| Descripción | texto | Sí | Hasta 1000 caracteres, sin espacios sobrantes |
| Severidad | `Crítica`, `Alta`, `Media` o `Baja` | Sí | Uno de esos cuatro valores |
| Sprint de detección | id de un sprint | Sí | Sprint **del proyecto** |
| Historia | id de una historia | No | Si viene, historia **del proyecto** |

## 3. Salidas esperadas

- El defecto queda guardado con su proyecto, descripción, severidad, sprint de detección, historia (si se eligió) y
  estado **Abierto**. Se vuelve a la lista de defectos del proyecto con el mensaje **"Defecto registrado."**.
- **Página de defectos** (`/proyectos/{id}/defectos`): los defectos del proyecto, el más nuevo primero, con
  severidad, estado, sprint de detección e historia; y el formulario. Se llega desde la vista del proyecto y desde la
  lista de proyectos.

## 4. Reglas de negocio

- RN1: La descripción es obligatoria, hasta 1000 caracteres, sin los espacios del inicio y del final (CA-27.1).
- RN2: La severidad es Crítica, Alta, Media o Baja (CA-27.1).
- RN3: Todo defecto nace en estado **Abierto** y sin sprint de resolución (CA-27.2). Los cambios de estado son de HU-28.
- RN4: El sprint de detección es obligatorio (la tabla lo pide) y tiene que ser del proyecto (CA-27.3).
- RN5: La historia es opcional y, si se elige, tiene que ser del proyecto (CA-27.3).
- RN6: Todos los errores de los campos se informan juntos.

## 5. Restricciones

- Reglas puras en el paquete nuevo `internal/domain/defect` (`Nuevo`); caso de uso `service.Defectos`; repositorios
  en memoria y PostgreSQL (`Guardar`, `ListarPorProyecto`).
- **Sin migración nueva:** la tabla `defectos` ya existe desde TEC-03.
- Un proyecto sin sprints no puede registrar defectos todavía: la página avisa que primero hay que crear un sprint.
- **Fuera de alcance:** cambiar el estado (HU-28), consultar con filtros (HU-29) y contar por sprint (HU-34).

## 6. Casos límite

| Caso | Resultado |
|---|---|
| "El login falla con Google", Alta, Sprint 1, HU-1 | Se registra Abierto |
| Sin historia | Se registra |
| Descripción vacía / de 1000 / de 1001 caracteres | Error / se registra / error |
| Severidad "Urgente" | Error |
| Sprint o historia de otro proyecto, o que no existen | Error |
| Sin sprint | Error: el sprint de detección es obligatorio |

## 7. Condiciones de error

| Situación | Campo | Error (Go) | Mensaje al usuario |
|---|---|---|---|
| Descripción vacía | descripcion | `ErrDescripcionObligatoria` | "la descripción del defecto es obligatoria" |
| Descripción larga | descripcion | `ErrDescripcionLarga` | "la descripción del defecto no puede superar los 1000 caracteres" |
| Severidad inválida | severidad | `ErrSeveridadInvalida` | "la severidad tiene que ser Crítica, Alta, Media o Baja" |
| Sin sprint o de otro proyecto | sprint | `ErrSprintInvalido` | "el sprint de detección tiene que ser un sprint del proyecto" |
| Historia de otro proyecto | historia | `ErrHistoriaInvalida` | "la historia tiene que ser del proyecto" |
| Proyecto inexistente | — | `project.ErrNoEncontrado` | Página 404 |

## 8. Criterios de aceptación

- CA-27.1: Descripción obligatoria y severidad Crítica, Alta, Media o Baja.
- CA-27.2: Arranca en estado Abierto.
- CA-27.3: La historia y el sprint tienen que ser del proyecto.

## Trazabilidad

Issue: #5 · Métrica: M8 · Feature: `features/hu-27-registrar-defecto.feature` · Pasos BDD: `features/hu27_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Dominio | `internal/domain/defect/defect.go` (`Nuevo`, severidades y estado Abierto) | `internal/domain/defect/defect_test.go` |
| Repositorios | `internal/store/memory/defectos.go`, `internal/store/postgres/defectos.go` | `defectos_test.go` de cada uno |
| Caso de uso | `internal/service/defectos.go` (`Ver`, `Registrar`) | `internal/service/defectos_test.go` |
| Pantalla | `internal/server/defectos.go`, `web/templates/paginas/defectos.html`, enlaces en `proyecto.html` y `proyectos.html` | `internal/server/defectos_test.go` |
| Aplicación | `app/app.go` | `app/app_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-27.1 Descripción y severidad | Registrar un defecto de una historia · Datos inválidos de un defecto | `TestNuevo_Validaciones`, `TestRegistrarDefecto_DatosInvalidos` |
| CA-27.2 Arranca Abierto | Registrar un defecto de una historia · Registrar un defecto sin historia | `TestNuevo_ArrancaAbierto`, `TestRegistrarDefecto_GuardaAbiertoYSeVeEnElProyecto` |
| CA-27.3 Historia y sprint del proyecto | El sprint y la historia tienen que ser del proyecto · Registrar un defecto sin historia | `TestNuevo_Validaciones`, `TestNuevo_SinHistoria`, `TestRegistrarDefecto_SprintDeOtroProyecto` |
