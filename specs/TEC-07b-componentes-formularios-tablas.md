# SPEC-TEC-07b · Componentes de formularios y tablas

**Tarea:** TEC-07b (#101, tareas #401 a #404) · **Responsable:** Bravo, Carolina · **Sprint:** 4 · **Estado:** Implementada (las demás pantallas adoptan los componentes cuando se toquen)

> Especificación escrita antes del código (SDD). Completa el CA-07.3 de TEC-07 (`specs/TEC-07-diseno-base.md`): la
> primera entrega (#99) trajo los mensajes, el error de campo y la ventana de confirmación. Sale de la revisión de
> Mariano en #99.

## 1. Objetivo

Que los formularios y las tablas de listas no repitan el mismo HTML en cada pantalla, y que en el celular ninguna tabla
deje un botón afuera de la pantalla (por ejemplo, el "Dar de baja" de integrantes).

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Datos del campo | `(dict "ID" … "Etiqueta" … …)` | Sí | `ID` y `Etiqueta` no vacíos |
| La tabla de una lista | `<table class="table … tabla-lista">` | No | Cada celda con `data-etiqueta` |

## 3. Salidas esperadas

- **Componente `campo`** (en `web/templates/componentes.html`):

```
{{template "campo" (dict "ID" "email" "Etiqueta" "Email" "Tipo" "email" "Valor" .Email
    "Error" (index .Errores "email") "Requerido" true "Max" 254 "Ayuda" "No puede repetirse dentro del proyecto.")}}
```

| Dato | Para qué | Si no está |
|---|---|---|
| `ID` | `id` y `name` del campo, y `for` de la etiqueta | — (obligatorio) |
| `Etiqueta` | El texto de la etiqueta | — (obligatorio) |
| `Tipo` | `type` del campo (`email`, `number`, `date`…) | Campo de texto |
| `Valor` | Lo que ya escribió la persona | Vacío |
| `Error` | El error debajo del campo, que lo marca en rojo (`is-invalid`) | Sin error |
| `Requerido` | `required` | No es obligatorio |
| `Max` | `maxlength` | Sin límite |
| `Foco` | `autofocus` | Sin foco |
| `Ayuda` | Texto de ayuda debajo, unido al campo con `aria-describedby` | Sin ayuda |

- **Estilo `tabla-lista`** (en `componentes.html`, bloque `estilos` que carga `base.html`): en pantallas de menos de
  576 px la tabla se apila; cada fila es un bloque y cada celda muestra su `data-etiqueta` antes del valor. La celda de
  acciones (`class="acciones"`) queda debajo, a lo ancho, así el botón siempre se ve. En la compu la tabla no cambia.

## 4. Reglas de negocio

- RN1: Un dato que no se pasa al componente no muestra nada (ni `<no value>`), como los componentes de TEC-07 (RN5).
- RN2: Con error, el campo tiene `is-invalid` y el mensaje va debajo (`errorCampo`).
- RN3: La ayuda se une al campo con `aria-describedby`, así un lector de pantalla la lee.
- RN4: En el celular, ninguna celda de una `tabla-lista` queda afuera de la pantalla.

## 5. Restricciones

- Sin librerías nuevas: Bootstrap 5.3 y un bloque `<style>` chico en `componentes.html`.
- `base.html` solo suma una línea (`{{template "estilos"}}`).
- Se adopta en las pantallas de integrantes (`integrantes.html` e `integrante_nuevo.html`), las de la observación de
  #99. Las demás pantallas lo adoptan cuando se toquen, para no pisar el trabajo de otros (Planning Poker está en
  revisión y en curso).

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Campo sin `Valor` | `value=""`, sin `<no value>` |
| Campo sin `Ayuda` | Sin `aria-describedby` ni texto de ayuda |
| Email largo en el celular | Se corta en varias líneas dentro del bloque |
| Integrante dado de baja | No tiene botón de baja (igual que antes) |

## 7. Condiciones de error

| Situación | Resultado |
|---|---|
| Una página usa `campo` sin `ID` | Campo sin `id`: los tests de la pantalla lo detectan |
| `dict` con datos impares | Error al mostrar la página (TEC-07) |

## 8. Criterios de aceptación (issue #101)

- CA-07b.1: Componente para un campo de formulario (etiqueta, campo, error debajo) que usen los formularios.
- CA-07b.2: Estilo común para las tablas de listas, que en el celular no deje botones fuera de pantalla.
- CA-07b.3: Las pantallas que ya existen los adoptan de a poco, cuando se toquen, sin pisar el trabajo de otros.

## Trazabilidad

Issue: #101 · Escenarios BDD: no aplica (no hay una regla de negocio nueva; se prueba con los tests de las pantallas,
igual que TEC-07).

| Tarea | Qué | Código | Tests |
|---|---|---|---|
| T1 (#401) | Componente de campo | `campo` en `web/templates/componentes.html` | `TestCampo_EtiquetaYAyudaUnidasAlCampo`, `TestCampo_ConErrorMarcaElCampoYMuestraElMensaje` |
| T2 (#402) | Tablas responsivas | `estilos` (`tabla-lista`) en `componentes.html`, cargado en `base.html` | `TestTablaLista_IntegrantesSeApilaEnElCelular` |
| T3 (#403) | Adopción | `integrante_nuevo.html` (`campo`, `error` y `errorCampo`) e `integrantes.html` (`tabla-lista`) | Los anteriores y los de HU-03 en `internal/server/integrantes_test.go` |
| T4 (#404) | PR y revisión | Este PR | — |

| Criterio | Tests |
|---|---|
| CA-07b.1 Campo de formulario | `TestCampo_EtiquetaYAyudaUnidasAlCampo`, `TestCampo_ConErrorMarcaElCampoYMuestraElMensaje` |
| CA-07b.2 Tablas en el celular | `TestTablaLista_IntegrantesSeApilaEnElCelular`; además lo revisé con un navegador de 375 px de ancho: el botón "Dar de baja" queda adentro de la pantalla |
| CA-07b.3 Adopción de a poco | Las pantallas de integrantes; las demás cuando se toquen |
