# SPEC-TEC-07 · Diseño base de la interfaz

**Tarea:** TEC-07 (#16) · **Responsable:** Bravo, Carolina · **Sprint:** 2 · **Estado:** Implementada (primera entrega; las demás páginas adoptan los componentes cuando se toquen)

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`. Es la primera entrega del diseño
> base: un PR chico el primer día del sprint, como pide la Planning del Sprint 2, para que las pantallas nuevas ya nazcan
> con él y no haya conflictos en `base.html`.

## 1. Objetivo

Que todas las pantallas se vean y se usen igual, en la compu y en el celular, y que el equipo no repita el mismo HTML en
cada página: el menú, los mensajes de confirmación y de error, el error debajo de cada campo y la confirmación antes de
una acción que no se puede deshacer.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| La página que se muestra | plantilla de `web/templates/paginas/` | Sí | Define `contenido` |
| Los componentes que usa | `{{template "nombre" …}}` | No | Están en `web/templates/componentes.html` |

## 3. Salidas esperadas

- **Layout base (`base.html`)**: un enlace "Ir al contenido" para quien navega con teclado; el menú que en el celular
  se junta en un botón (☰) y en la compu se ve completo; el contenido; y un pie con el nombre de la app.
- **Componentes (`componentes.html`)**, para usar desde cualquier página:

| Componente | Uso | Qué muestra |
|---|---|---|
| `exito` | `{{template "exito" "Proyecto creado."}}` | Mensaje verde que se puede cerrar (`role="status"`) |
| `error` | `{{template "error" .ErrorGeneral}}` | Mensaje rojo (`role="alert"`); si el texto está vacío no muestra nada |
| `aviso` | `{{template "aviso" .Error}}` | Mensaje amarillo (`role="alert"`); si está vacío no muestra nada |
| `errorCampo` | `{{template "errorCampo" (index .Errores "nombre")}}` | El error debajo del campo; si no hay error no muestra nada |
| `confirmar` | `{{template "confirmar" (dict "ID" … "Titulo" … "Texto" … "Accion" … "Boton" …)}}` | Un botón que abre una ventana (modal) para confirmar; recién al confirmar se envía el formulario |

## 4. Reglas de negocio

- RN1: El menú se colapsa en pantallas chicas (menos de 768 px) y se abre con el botón ☰.
- RN2: Los mensajes de éxito se pueden cerrar; los de error y aviso no, para que no se pierdan.
- RN3: Los mensajes de error y aviso usan `role="alert"` y los de éxito `role="status"`, así un lector de pantalla los anuncia.
- RN4: Una acción que no se puede deshacer (por ejemplo dar de baja a un integrante) pide confirmación en una ventana
  de la página y no con el cuadro del navegador (`confirm()`), que no se ve igual en todos lados.
- RN5: Un componente con el texto vacío no muestra nada, así las páginas no necesitan `{{if}}` alrededor.

## 5. Restricciones

- Bootstrap 5.3 (ya estaba): se agrega su JavaScript (`bootstrap.bundle.min.js`) porque el menú y las ventanas lo usan.
- Los componentes se cargan junto con `base.html` en cada página (`internal/server/render.go`), y se agrega la función
  `dict` para pasarle varios datos a un componente.
- Esta entrega cambia lo mínimo en las páginas: usa los componentes en las listas de proyectos y de integrantes y en el
  formulario de nuevo proyecto. Las demás páginas los adoptan cuando se toquen, para no pisar el trabajo de otros.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Pantalla de celular | El menú queda detrás del botón ☰ |
| `{{template "error" ""}}` | No muestra nada |
| `{{template "errorCampo" nil}}` | No muestra nada |
| Integrante dado de baja | No tiene botón ni ventana de baja |
| Cancelar en la ventana de confirmación | No se envía nada |

## 7. Condiciones de error

| Situación | Resultado |
|---|---|
| Una página usa un componente que no existe | La plantilla falla al cargar y los tests lo detectan antes de mergear |
| `dict` con una cantidad impar de datos o una clave que no es texto | Error al mostrar la página (no pasa con los componentes de esta spec) |

## 8. Criterios de aceptación (issue #16)

- CA-07.1: `base.html` con menú, mensajes de error y confirmaciones.
- CA-07.2: Se usa bien en celular y en compu.
- CA-07.3: Componentes reutilizables (formularios, tablas, ventanas).

## Trazabilidad

Issue: #16 · Escenarios BDD: no aplica (no hay una regla de negocio nueva; el comportamiento se prueba con los tests de las
pantallas, igual que en TEC-03).

| Parte | Código | Tests |
|---|---|---|
| Layout | `web/templates/base.html` | `internal/server/layout_test.go` |
| Componentes | `web/templates/componentes.html` (`exito`, `error`, `aviso`, `errorCampo`, `confirmar`) | `internal/server/componentes_test.go` |
| Carga y `dict` | `internal/server/render.go` | `internal/server/dict_test.go` |
| Páginas que ya los usan | `proyectos.html`, `integrantes.html`, `proyecto_nuevo.html` | los tests de cada pantalla siguen en verde |

| Criterio | Tests |
|---|---|
| CA-07.1 Menú, errores y confirmaciones | `TestLayout_MenuSeColapsaEnElCelular`, `TestMensajeDeExito_SeAnunciaYSePuedeCerrar`, `TestAviso_NoSePuedeCerrar`, `TestDarDeBaja_PideConfirmacionEnUnaVentana` |
| CA-07.2 Celular y compu | `TestLayout_MenuSeColapsaEnElCelular`, `TestLayout_EnlaceParaIrAlContenidoYPie` |
| CA-07.3 Componentes reutilizables | `TestMensajeDeExito_*`, `TestAviso_*`, `TestDarDeBaja_*`, `TestDict` |
