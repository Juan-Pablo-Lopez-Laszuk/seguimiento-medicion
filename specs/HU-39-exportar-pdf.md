# SPEC-39 · Exportar reporte a PDF

**Historia:** HU-39 (#50) · **Responsable:** López, Juan Pablo · **Sprint:** 4 · **Estado:** Revisada

> Especificación escrita antes del código (SDD). Pasa a PDF los dos reportes que ya existen: el de un sprint (HU-37) y
> el del proyecto (HU-38). No calcula nada nuevo: el PDF se arma con los mismos datos y los mismos textos que la
> pantalla, así no pueden dar distinto (CA-39.2).
> Spike del 24/10 (#375): `codeberg.org/go-pdf/fpdf` v0.11.1 (Go puro, licencia MIT, fuentes base incluidas) genera
> 12 páginas con tablas en unos 20 ms. La v0.12 pide Go 1.25 y el módulo está en 1.24, así que va la v0.11.1.
> Revisada por Juan Pablo el 24/10: los dos PDF (sprint y proyecto), el nombre con la fecha AAAA-MM-DD, las métricas
> en tabla, el pie con la fecha y la página, y `?` en lugar de los caracteres que el PDF no tiene.

## 1. Objetivo

Que el Agile Enabler pueda descargar el reporte de un sprint o del proyecto como un archivo PDF, para mandárselo al
cliente o a los profesores sin que tengan que entrar al sistema.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Sprint | número del sprint (viene en la URL, solo en el PDF de un sprint) | Sí | Tiene que existir en el proyecto |

No hay datos que cargue el usuario. Hay dos descargas, cada una con un botón **Descargar PDF** al lado de **Imprimir**:

| Reporte | Página (HU-37 y HU-38) | Descarga |
|---|---|---|
| De un sprint | `/proyectos/{id}/sprints/{sprint}/reporte` | `/proyectos/{id}/sprints/{sprint}/reporte.pdf` |
| Del proyecto | `/proyectos/{id}/reporte` | `/proyectos/{id}/reporte.pdf` |

## 3. Salidas esperadas

Una respuesta `200` con `Content-Type: application/pdf` y `Content-Disposition: attachment`, para que el navegador
descargue el archivo en lugar de abrirlo.

**Nombre del archivo (CA-39.1):** `proyecto-sprint-fecha.pdf`.
- **proyecto:** el nombre del proyecto en minúsculas, sin tildes, con guiones en lugar de espacios y otros signos.
- **sprint:** `sprint-N`. El PDF del proyecto no lo lleva.
- **fecha:** el día en que se genera, en la Argentina (`domain.HoyEnArgentina`), con el formato `AAAA-MM-DD` para que
  los archivos queden ordenados.

Con el ejemplo de la sección 6, generado el 29/10/2026:

| Reporte | Archivo |
|---|---|
| Sprint 2 | `software-metrics-sprint-2-2026-10-29.pdf` |
| Proyecto | `software-metrics-2026-10-29.pdf` |

**Contenido (CA-39.2):** lo mismo que la pantalla, en el mismo orden y con los mismos textos.

- **PDF de un sprint:**
  1. el título "Reporte del Sprint 2 · Software Metrics", el objetivo, las fechas y el estado;
  2. el aviso de reporte parcial si el sprint no está cerrado (HU-37 RN2);
  3. las métricas en una tabla de dos columnas, métrica y valor, con los siete valores de HU-37;
  4. la tabla de historias (ID, título, SP, resultado, horas estimadas, reales y desviación);
  5. la tabla de defectos, o "No hubo defectos en este sprint.".
- **PDF del proyecto:**
  1. el título "Reporte del proyecto · Software Metrics", la descripción, las fechas y el estado;
  2. el resumen del proyecto (los seis valores de HU-38);
  3. la evolución de la velocidad;
  4. un bloque por sprint, cada uno en una hoja nueva, con el mismo contenido que el PDF de ese sprint.

Cada hoja lleva al pie "Generado el 29/10/2026 · Página 1 de 4". El PDF no tiene la navegación, los botones ni los
enlaces de la pantalla.

## 4. Reglas de negocio

- RN1: **El PDF muestra lo mismo que la pantalla** (CA-39.2): se arma con el mismo `ReporteSprint` o `ReporteProyecto`
  y con los mismos métodos de texto (`SPTexto`, `CumplimientoTexto`, etc.). Si cambia un texto en la pantalla, cambia
  en el PDF.
- RN2: **Mismas reglas que el reporte:** foto del cierre en los sprints cerrados y aviso en los abiertos (HU-37), y
  totales del proyecto que salen del panel y no de sumar los bloques (HU-38 RN2).
- RN3: **Nombre del archivo** según la sección 3. Si al nombre del proyecto no le queda ninguna letra ni número (por
  ejemplo "¿¡!?"), se usa `proyecto`.
- RN4: **Los caracteres que las fuentes del PDF no tienen** (emojis, flechas, letras de otros alfabetos) se reemplazan
  por `?`. El castellano sale completo: tildes, ñ, ü, ¿, ¡, «», · y €.
- RN5: **Un título o una descripción larga** se parte en varias líneas dentro de su celda; no se corta.
- RN6: Es solo lectura: descargar el PDF no cambia nada.

## 5. Restricciones

- **Menos de 10 segundos (CA-39.3)**, el límite de la función de Vercel. El PDF se genera en el servidor, en memoria y
  sin servicios externos. El spike dio 20 ms para 12 páginas.
- **Librería:** `codeberg.org/go-pdf/fpdf` v0.11.1, con la fuente Helvetica que trae incluida (no hay archivos de
  fuentes que desplegar).
- **Sin cálculos nuevos, sin migración ni repositorios nuevos.** El caso de uso es el de HU-37 y HU-38
  (`Reportes.Sprint` y `Reportes.Proyecto`).
- **Capas:**
  - El paquete nuevo `internal/reportepdf` es de presentación, como las plantillas HTML.
  - `DocumentoSprint(ReporteSprint)` y `DocumentoProyecto(ReporteProyecto)` arman un documento con títulos, párrafos,
    avisos, tablas y saltos de hoja. Son funciones puras, sin la librería, y ahí se prueba que el contenido es el de la
    pantalla.
  - `Escribir(w, documento)` lo dibuja con fpdf.
  - Los handlers de `internal/server/reportes.go` arman el nombre del archivo y escriben la respuesta.
- **Depende de HU-38 (#423):** la rama sale de la de HU-38 y se rebasa sobre `main` cuando se mergee.

## 6. Casos límite

Mismo proyecto que la spec de HU-38: "Software Metrics" (05/10 al 30/11) con el Sprint 1 cerrado (5 de 13 SP), el
Sprint 2 cerrado (3 de 8 SP) y el Sprint 3 abierto; tres historias con horas y tres defectos. Se genera el 29/10/2026.

| Caso | Resultado |
|---|---|
| PDF del Sprint 2 | `software-metrics-sprint-2-2026-10-29.pdf` con las métricas, las historias y los defectos del reporte del Sprint 2 |
| PDF del Sprint 3 (abierto) | Lleva el aviso "El sprint no está cerrado: el reporte muestra cómo va hasta hoy." |
| PDF del proyecto | `software-metrics-2026-10-29.pdf`: resumen, evolución y tres bloques, cada uno en una hoja nueva |
| Sprint sin historias ni defectos | "El sprint no tiene historias." y "No hubo defectos en este sprint.", como en la pantalla |
| Proyecto sin sprints | Resumen con "sin datos", "Aún no hay sprints cerrados" y "Todavía no hay sprints en este proyecto." |
| Historia con un título de 200 caracteres | El título se parte en varias líneas dentro de la celda (RN5) |
| Historia "Login 🔐 → OAuth" | En el PDF sale "Login ? ? OAuth" (RN4) |
| Proyecto "Gestión Ágil 2026" | El archivo empieza con `gestion-agil-2026-` |
| Proyecto con 12 sprints de 30 historias | Se genera en menos de 1 segundo en la máquina de desarrollo (margen sobre los 10 s de Vercel) |

## 7. Condiciones de error

| Situación | Error | Mensaje al usuario |
|---|---|---|
| El proyecto no existe | 404 | "No encontramos ese proyecto." (el de HU-11) |
| El sprint no existe en el proyecto | 404 | "No encontramos ese sprint." (el de HU-37) |
| Falla al leer la base o al generar el PDF | 500 | — |

## 8. Criterios de aceptación

- CA-39.1 Descarga un .pdf llamado proyecto-sprint-fecha.pdf.
- CA-39.2 El contenido coincide con el reporte en pantalla.
- CA-39.3 Se genera en menos de 10 segundos (límite de la función serverless).

## Trazabilidad

Issue: #50 · Feature: `features/hu-39-exportar-pdf.feature` · Pasos BDD: `features/hu39_steps_test.go`

Las tablas por capa y por criterio se completan con el código.
