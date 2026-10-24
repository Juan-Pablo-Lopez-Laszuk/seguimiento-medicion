# Guion de la presentación final

> **Borrador** (24/10/2026, Juan Pablo). Entregable 13 de la consigna: presentación formal del producto al cliente en
> la Sprint Review final (lunes 02/11). Lo marcado [A CONFIRMAR] lo cerramos en el ensayo del sábado 31/10 o el
> domingo 01/11.

## Qué evalúan y cómo lo cubrimos

| Criterio (peso) | Qué mostramos | Quién |
|---|---|---|
| Producto funcional (25 %) | La demo en producción, de punta a punta, con un proyecto de ejemplo | Los tres, cada uno sus épicas |
| SDD, BDD y TDD (25 %) | El recorrido de trazabilidad de una funcionalidad: historia → spec → criterios → escenarios → tests → código | El dueño de la historia que elijan |
| Calidad del software (20 %) | La arquitectura en capas, el manejo de errores, el CI y la cobertura | Juan Pablo y Mariano |
| Gestión del proyecto (20 %) | El tablero, las plannings, las reviews, las retros y el historial de Git | Juan Pablo |
| Trabajo en equipo y presentación (10 %) | Hablan los tres y cada uno justifica sus decisiones | Los tres |

Los profesores eligen **al menos una funcionalidad** para el recorrido de trazabilidad. Por eso cada uno ensaya el
recorrido de una historia suya (sección 3).

## 1. Orden y tiempos

[A CONFIRMAR] La duración con los profesores. Esta propuesta suma unos 25 minutos.

| # | Parte | Min | Quién |
|---|---|---|---|
| 1 | El problema y el producto: qué resuelve, para quién, cómo se usa | 2 | Juan Pablo |
| 2 | Cómo trabajamos: 5 sprints, tablero, rotación de revisiones, SDD → BDD → TDD, uso de IA registrado | 3 | Juan Pablo |
| 3 | Demo, parte 1: login, proyecto, integrantes, backlog y Planning Poker | 5 | Mariano y Juan Pablo |
| 4 | Demo, parte 2: sprint, tareas, horas, defectos, cierre | 5 | Juan Pablo y Carolina |
| 5 | Demo, parte 3: panel, gráficos, reportes y PDF | 4 | Carolina y Juan Pablo |
| 6 | Trazabilidad de punta a punta de la funcionalidad que elijan | 4 | El dueño de la historia |
| 7 | Calidad: arquitectura, CI, cobertura, matriz de trazabilidad | 2 | Mariano |
| — | Preguntas | — | Los tres |

## 2. La demo, paso a paso

Se hace en **producción** (https://seguimiento-medicion.vercel.app) con el proyecto de ejemplo ya cargado (sección 4).

1. **Login** (TEC-05). Mariano entra con su cuenta. Muestra que sin sesión todas las páginas piden ingresar.
2. **Proyecto** (HU-01 a HU-04). Juan Pablo muestra la página del proyecto: estado calculado, sprint activo, días
   restantes y SP completados. Intenta crear un proyecto con la fecha de fin antes del inicio, para mostrar los errores
   por campo.
3. **Backlog** (HU-05 a HU-10). Mariano crea una historia con un criterio, la prioriza con Subir y Bajar y filtra el
   backlog.
4. **Planning Poker** (HU-17 a HU-22). Mariano abre la sesión para la historia nueva. Carolina y Juan Pablo votan desde
   el celular, se revela, se ve la diferencia y se registra la estimación acordada.
5. **Sprint** (HU-11 a HU-13). Juan Pablo asigna la historia al sprint activo (aparece el aviso de velocidad si se pasa),
   la empieza y la completa con los criterios cumplidos.
6. **Esfuerzo y defectos** (HU-23 a HU-29). Carolina divide la historia en tareas, registra horas y carga un defecto
   con su severidad.
7. **Cerrar el sprint** (HU-14 y HU-15). Juan Pablo cierra el sprint: la foto del cierre, las historias que vuelven al
   backlog y el cumplimiento en la lista de sprints.
8. **Métricas** (HU-30 a HU-36). Carolina muestra el panel, los gráficos y las horas con su desviación.
9. **Reportes** (HU-37 a HU-39). Juan Pablo abre el reporte del proyecto, la evolución de la velocidad y descarga el
   PDF.

**Plan B:** si producción no responde, la misma demo en la app local (`go run ./cmd/server`, en modo memoria). El modo
memoria arranca vacío y el repo no tiene un script de datos de ejemplo. [A CONFIRMAR] Armar uno que cargue el proyecto de
ejemplo por la misma interfaz y dejarlo probado en las tres computadoras.

## 3. Recorrido de trazabilidad (ensayar una historia cada uno)

Para la historia que elijan los profesores, se muestra en este orden:

1. **Historia** en el tablero: el issue con sus criterios de aceptación y sus tareas con horas.
2. **Spec SDD** en `specs/`: reglas de negocio, casos límite, errores y la tabla de trazabilidad por capa y por
   criterio.
3. **Escenarios BDD** en `features/`: cada escenario con la etiqueta del criterio que cubre (`@CA-39.1`) y los pasos
   en Go.
4. **Tests y TDD**: el commit `test(HU-xx): RED` antes del `feat(HU-xx): GREEN` en el historial del PR. El test falla
   en el RED y pasa en el GREEN.
5. **Código Go**: el caso de uso en `internal/service` o el dominio, y la pantalla en `internal/server`.
6. **Matriz de la consigna**: la fila del requerimiento en `docs/trazabilidad-requisitos.md`, que un test del CI
   mantiene al día.

| Integrante | Historia para ensayar | Por qué |
|---|---|---|
| Juan Pablo | HU-39 · Exportar reporte a PDF | Los pasos BDD comparan el PDF descargado con la página; tiene un spike documentado |
| Mariano | HU-18 · Votar de forma oculta [A CONFIRMAR] | Tiene reglas de seguridad (la API no expone los votos) y casos de error |
| Carolina | HU-35 · Panel de indicadores | Junta varias métricas y tiene la regla de "sin datos" |

## 4. Preparación (antes del 02/11)

- [ ] Proyecto de ejemplo cargado en producción con 3 sprints (2 cerrados), historias con horas y defectos, para que el
      panel, los gráficos y los reportes tengan datos. [A CONFIRMAR] Quién lo carga y cuándo.
- [ ] Cuentas de los profesores creadas en Supabase (el registro está cerrado), por si quieren probar.
- [ ] Los tres con su cuenta vinculada a su integrante en el proyecto de ejemplo, para votar en Planning Poker.
- [ ] Pestañas abiertas: tablero, el PR de la historia de trazabilidad, la spec, el `.feature` y la matriz.
- [ ] Ensayo completo con tiempos el sábado 31/10 o el domingo 01/11.
- [ ] Plan B probado.

## 5. Preguntas que nos pueden hacer

| Pregunta | Quién responde | Respuesta corta |
|---|---|---|
| ¿Por qué páginas en el servidor y no React? | Juan Pablo | Un solo lenguaje, sin build de frontend; la consigna pide el núcleo en Go (decisión 1 de `docs/arquitectura.md`) |
| ¿Cómo garantizan que el reporte y el panel no den distinto? | Juan Pablo | Las métricas son funciones puras en `internal/metrics` y se reutilizan; el reporte del proyecto toma el resumen del panel |
| ¿Qué pasa si se re-estima una historia de un sprint cerrado? | Juan Pablo | No cambia: el sprint guarda la foto del cierre (HU-14) |
| ¿Cómo validan el login sin llamar a Supabase en cada pedido? | Mariano | Con las claves públicas del JWKS, localmente (TEC-05) |
| ¿Cómo saben que la cobertura no baja? | Mariano | El CI falla si `internal/domain` e `internal/metrics` bajan del 80 % (TEC-06) |
| ¿Cómo calculan la desviación de horas? | Carolina | Reales menos estimadas, sobre las estimadas; sin estimadas no aplica (HU-33) |
| ¿Cómo usaron la IA? | Los tres | Para specs, escenarios y código, siempre revisado y con tests; registrado en `docs/ia/registro.md` |
| ¿Qué no llegaron a hacer? | Juan Pablo | [A COMPLETAR después del congelamiento] Por ejemplo, permisos por rol |
