# language: es
@HU-34
Característica: Defectos detectados y resueltos
  Como Agile Enabler
  Quiero saber cuántos defectos se detectaron y resolvieron
  Para medir la calidad

  Antecedentes:
    Dado el proyecto con los sprints 1 y 2 y estos defectos:
      | severidad | detectado en | estado    | resuelto en |
      | Alta      | Sprint 1     | Resuelto  | Sprint 1    |
      | Alta      | Sprint 1     | Cerrado   | Sprint 2    |
      | Baja      | Sprint 1     | Abierto   |             |
      | Crítica   | Sprint 2     | Reabierto |             |

  # Caso NORMAL (CA-34.1)
  @CA-34.1
  Escenario: Detectados y resueltos por sprint
    Cuando veo el resumen de defectos con todas las severidades
    Entonces el Sprint 1 tiene 3 detectados y 1 resuelto
    Y el Sprint 2 tiene 1 detectado y 1 resuelto
    Y el proyecto tiene 4 detectados, 2 resueltos y 2 abiertos

  # Caso ALTERNATIVO (CA-34.2): filtro por severidad
  @CA-34.2
  Escenario: Filtrar por severidad
    Cuando veo el resumen de defectos de severidad "Alta"
    Entonces el Sprint 1 tiene 2 detectados y 1 resuelto
    Y el Sprint 2 tiene 0 detectados y 1 resuelto
    Y el proyecto tiene 2 detectados, 2 resueltos y 0 abiertos

  # Caso LÍMITE: un reabierto no cuenta como resuelto y una severidad desconocida muestra todas
  @CA-34.1 @CA-34.2
  Escenario: Severidad desconocida
    Cuando veo el resumen de defectos de severidad "Urgente"
    Entonces el proyecto tiene 4 detectados, 2 resueltos y 2 abiertos
