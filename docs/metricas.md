# Métricas del sistema

> **Proyecto:** Software Metrics & Estimation · TPI Ingeniería y Calidad de Software 2026 · UTN FRSR
> **Responsable:** Bravo, Carolina · Épicas E5 (Esfuerzo), E6 (Defectos), E7 (Métricas), E8 (Dashboard)
> **Estado:** v1.0 (acordado en la Review del Sprint 0) · Las preguntas abiertas que afectaban a las métricas quedaron resueltas (sección 8).
> **Uso de IA:** borrador elaborado con asistencia de IA, en revisión por la responsable; ver `docs/ia/registro.md`.

## 1. Para qué sirve este documento

Este documento es el **glosario compartido del equipo** sobre métricas. Define, **antes de programar**, qué significa cada métrica, cómo se calcula y qué pasa en los casos raros.

Lo usa todo el equipo:

- **Carolina**: para escribir las especificaciones (SDD) y los tests (TDD) de las métricas y del dashboard.
- **Juan Pablo**: los reportes reutilizan estos mismos cálculos, así no se repiten fórmulas.
- **Mariano**: para saber qué datos tiene que guardar la base de datos.

La consigna del TPI pide como mínimo nueve métricas (punto 7 del enunciado). Todas están definidas en la sección 4.

## 2. Reglas generales (aplican a todas las métricas)

| # | Regla |
|---|---|
| RG-1 | Los cálculos son **funciones puras** en `internal/metrics`: reciben datos simples (structs) y devuelven un resultado. No acceden a la base de datos ni a HTTP. Así se testean con TDD sin depender de nadie. |
| RG-2 | Internamente se calcula con `float64`. **En pantalla se redondea a 1 decimal** (ej.: 16,5). |
| RG-3 | **Nunca se divide por cero.** Cada métrica define qué devuelve cuando el divisor es 0. |
| RG-4 | Los datos inválidos (story points u horas negativas) devuelven un **error con nombre propio** (`ErrSPNegativos`, etc.), no un número equivocado. |
| RG-5 | Cada métrica se puede calcular **por sprint** y **por proyecto** (proyecto = todos sus sprints), salvo que se indique otra cosa. |
| RG-6 | Para un **sprint cerrado** se usan los valores guardados al cerrarlo ("foto del cierre", HU-14 CA-14.3). Para el **sprint activo**, se calculan en el momento. |

## 3. Conceptos que usan las fórmulas

| Concepto | Significado |
|---|---|
| **Story Points (SP)** | Número que expresa el esfuerzo relativo de una historia. Escala Fibonacci: 0, 1, 2, 3, 5, 8, 13, 21 (sin "?": un voto siempre es un número de la escala). |
| **Historia Hecha** | Historia en estado `Hecho`. Solo pasa a Hecho si cumple todos sus criterios de aceptación (HU-08). |
| **Estados de un sprint** | `Planificado` → `Activo` → `Cerrado`. |
| **Registro de esfuerzo** | Integrante + fecha + actividad + horas trabajadas en una historia o tarea (HU-24). |
| **Defecto resuelto** | Defecto en estado `Resuelto` o `Cerrado` que tiene sprint de resolución. Si se reabre, deja de contar como resuelto (HU-28 CA-28.3). |

## 4. Métricas obligatorias de la consigna

### M1 · Story Points planificados — HU-30

- **Qué mide:** cuánto trabajo se comprometió el equipo en un sprint.
- **Fórmula:** `SP planificados = Σ SP de las historias asignadas al sprint`
- **Proyecto:** suma de los SP planificados de todos sus sprints.
- **Casos límite:** sprint sin historias → 0. Una historia sin estimar suma 0 y **se muestra una advertencia** ("hay N historias sin estimar", CA-30.2).
- **Errores:** SP negativos → `ErrSPNegativos`.

### M2 · Story Points completados — HU-30

- **Qué mide:** cuánto trabajo se terminó realmente.
- **Fórmula:** `SP completados = Σ SP de las historias del sprint en estado Hecho`
- **Proyecto:** suma de los SP completados de todos sus sprints.
- **Casos límite:** ninguna historia Hecha → 0. Las historias que vuelven al backlog al cerrar el sprint no suman como completadas (HU-14 CA-14.2).
- **Errores:** SP negativos → `ErrSPNegativos`.

### M3 · Velocidad del equipo — HU-31

- **Qué mide:** cuántos SP termina el equipo por sprint, en promedio. Sirve para decidir cuánto comprometer en el próximo sprint.
- **Fórmula:** `Velocidad = (Σ SP completados de los últimos N sprints cerrados) / k`, donde k es la cantidad de sprints que se tomaron: k = N si hay N o más sprints cerrados; si hay menos, k = los que haya (ver el ejemplo de la sección 6).
- **Ventana N:** por defecto **3** sprints, configurable por proyecto. "Últimos" son los de mayor número de sprint.
- **Reglas:** solo cuentan sprints **Cerrados**; los activos o planificados no (CA-31.4). El orden en que llegan los datos no cambia el resultado.
- **Casos límite:** sin sprints cerrados → 0 y el mensaje "Aún no hay sprints cerrados" (CA-31.3). Si hay menos de N sprints cerrados, se promedian los que haya (CA-31.2).
- **Errores:** N ≤ 0 → `ErrVentanaInvalida` (CA-31.5). SP negativos → `ErrSPNegativos`.
- **Ejemplo:** sprints cerrados con 10, 20, 30 y 40 SP; N = 3 → (20 + 30 + 40) / 3 = **30**.

### M4 · Horas estimadas — HU-23

- **Qué mide:** cuánto tiempo se esperaba dedicar.
- **Fórmula (historia):** si la historia tiene tareas, `Horas estimadas = Σ horas estimadas de sus tareas`. Si no tiene tareas, se usa la estimación cargada en la historia (solo se carga si la historia no tiene tareas; ver `docs/modelo-datos.md`).
- **Sprint / proyecto:** suma de las horas estimadas de sus historias.
- **Validaciones:** mayor que 0; admite decimales (ej.: 1,5).
- **Errores:** horas ≤ 0 al cargar → `ErrHorasInvalidas`.

### M5 · Horas reales — HU-24

- **Qué mide:** cuánto tiempo se trabajó de verdad.
- **Fórmula:** `Horas reales = Σ horas de los registros de esfuerzo` (de la historia, del sprint o del proyecto).
- **Validaciones de cada registro (HU-24):** horas > 0 y ≤ 24. El total diario de un integrante no supera 24 h. La fecha no es futura y está dentro del proyecto. El integrante pertenece al proyecto.
- **Casos límite:** sin registros → 0.
- **Errores:** `ErrHorasInvalidas`, `ErrExcedeHorasDiarias`, `ErrFechaFutura`, `ErrFechaFueraDeProyecto`, `ErrIntegranteAjeno`.

### M6 · Desviación entre esfuerzo estimado y real — HU-33

- **Qué mide:** cuánto se equivocó la estimación de horas.
- **Fórmulas:**
  - `Desviación (h) = Horas reales − Horas estimadas`
  - `Desviación (%) = (Horas reales − Horas estimadas) / Horas estimadas × 100`
- **Cómo se lee:** positivo = se tardó **más** de lo estimado; negativo = se tardó **menos**.
- **Alcance:** por historia, por sprint y por proyecto (CA-33.3).
- **Casos límite:** horas estimadas = 0 → la desviación % se informa como **"no aplicable"**, no como error (CA-33.2).
- **Ejemplo:** estimadas 10 h, reales 12 h → +2 h y **+20 %**.

### M7 · Porcentaje de historias completadas — HU-32

- **Qué mide:** qué parte de lo comprometido se terminó, contando **historias** (no SP).
- **Fórmula:** `% completadas = (historias Hechas / historias planificadas) × 100`
- **Alcance:** por sprint y por proyecto (CA-32.3). Se muestra con 1 decimal.
- **Casos límite:** sprint sin historias → **0 %, sin error** (división por cero controlada, CA-32.2).
- **Ejemplo:** 3 de 4 historias Hechas → **75,0 %**.

### M8 · Cantidad de defectos detectados — HU-34

- **Qué mide:** cuántos errores se encontraron.
- **Fórmula:** `Detectados (sprint S) = cantidad de defectos cuyo sprint de detección es S`
- **Proyecto:** total de defectos registrados. Se puede filtrar por severidad (Crítica, Alta, Media, Baja).
- **Casos límite:** sin defectos → 0.

### M9 · Cantidad de defectos resueltos — HU-34

- **Qué mide:** cuántos errores se corrigieron.
- **Fórmula:** `Resueltos (sprint S) = cantidad de defectos en estado Resuelto o Cerrado cuyo sprint de resolución es S`
- **Proyecto:** total de defectos resueltos. Filtro por severidad.
- **Reglas:** un defecto reabierto no cuenta como resuelto (se limpia su sprint de resolución). El sprint de resolución es igual o posterior al de detección (HU-28 CA-28.2).
- **Casos límite:** sin defectos resueltos → 0.

## 5. Métricas derivadas para el dashboard (HU-35, HU-36)

| Métrica | Fórmula | Uso |
|---|---|---|
| Defectos abiertos | Detectados − Resueltos (o conteo de defectos que no están Resueltos ni Cerrados) | Tarjeta del dashboard |
| Burndown del sprint activo | SP pendientes al final de cada día = SP planificados − SP completados hasta ese día | Gráfico del sprint activo |
| Precisión de estimación *(valor agregado)* | SP completados / SP planificados × 100 | Informe final |

## 6. Ejemplo completo (sirve para tests y para la demo)

Proyecto con 2 sprints cerrados:

| Sprint | Historias planificadas | SP planificados | Historias Hechas | SP completados |
|---|---|---|---|---|
| 1 | 4 | 20 | 3 | 15 |
| 2 | 3 | 18 | 3 | 18 |

Historia "Login": estimada en 10 h; registros de 4 h, 5 h y 3 h. Defectos: 2 detectados en el sprint 1, 1 resuelto en el sprint 1.

| Métrica | Resultado |
|---|---|
| SP planificados (proyecto) | 20 + 18 = **38** |
| SP completados (proyecto) | 15 + 18 = **33** |
| Velocidad (N = 3, hay 2 cerrados) | (15 + 18) / 2 = **16,5** |
| % completadas sprint 1 | 3 / 4 × 100 = **75,0 %** |
| % completadas proyecto | 6 / 7 × 100 = **85,7 %** |
| Horas reales "Login" | 4 + 5 + 3 = **12 h** |
| Desviación "Login" | 12 − 10 = **+2 h (+20,0 %)** |
| Defectos detectados / resueltos / abiertos | **2 / 1 / 1** |

## 7. Trazabilidad

| Métrica | Historia | Spec SDD | Escenarios BDD | Código (propuesto) |
|---|---|---|---|---|
| M1, M2 | HU-30 | `specs/HU-30-sp-planificados-completados.md` | `features/hu-30-sp.feature` | `internal/metrics/storypoints.go` |
| M3 | HU-31 | `specs/HU-31-velocidad.md` | `features/hu-31-velocidad.feature` | `internal/metrics/velocity.go` |
| M4 | HU-23 | `specs/HU-23-horas-estimadas.md` | `features/hu-23-horas-estimadas.feature` | `internal/domain/effort/` |
| M5 | HU-24 | `specs/HU-24-registro-esfuerzo.md` | `features/hu-24-registro-esfuerzo.feature` | `internal/domain/effort/` + `internal/metrics/hours.go` |
| M6 | HU-33 | `specs/HU-33-desviacion.md` | `features/hu-33-desviacion.feature` | `internal/metrics/deviation.go` |
| M7 | HU-32 | `specs/HU-32-porcentaje-completadas.md` | `features/hu-32-porcentaje.feature` | `internal/metrics/completion.go` |
| M8, M9 | HU-34 | `specs/HU-34-defectos.md` | `features/hu-34-defectos.feature` | `internal/metrics/defects.go` |

## 8. Preguntas que afectaban a las métricas (resueltas)

Se cerraron en la Review del Sprint 0 (`docs/scrum/sprint-0/review.md`).

| # (plan) | Pregunta | Decisión | Quién la resolvió |
|---|---|---|---|
| 7 | ¿Qué escala de story points usamos? | Fibonacci 0, 1, 2, 3, 5, 8, 13, 21, sin "?" | Equipo |
| 9 | ¿Horas estimadas por historia, por tarea o ambas? | Ambas: la historia suma las horas de sus tareas | Equipo |
| 10 | ¿Sobre cuántos sprints se calcula la velocidad? | Últimos 3 sprints cerrados | Equipo |
| 13 | ¿Cobertura mínima de tests exigida? | La cátedra no fija un porcentaje: hay que cubrir todo lo que pide la consigna. El 80 % en `internal/metrics` e `internal/domain` es meta interna del equipo | Profesores y equipo |
