# language: es
@HU-26
Característica: Consultar y corregir registros de esfuerzo
  Como integrante
  Quiero ver y corregir mis horas cargadas
  Para que los datos sean confiables

  Antecedentes:
    Dado el proyecto con el Sprint 1 del 05/10/2026 al 18/10/2026, el Sprint 2 del 19/10/2026 al 01/11/2026 y hoy 21/10/2026
    Y el equipo con "Juan Pablo" de Agile Enabler y "Carolina" y "Mariano" de Product Builder
    Y las historias con tareas estimadas:
      | historia | título   | horas estimadas |
      | HU-1     | Login    | 6               |
      | HU-2     | Reportes | 4               |
    Y estos registros de horas:
      | integrante | historia | fecha      | actividad       | horas |
      | Carolina   | HU-1     | 2026-10-10 | Pantalla        | 3     |
      | Mariano    | HU-1     | 2026-10-12 | Validaciones    | 2     |
      | Carolina   | HU-2     | 2026-10-20 | Primer reporte  | 4     |

  # Caso NORMAL (CA-26.1): los filtros se combinan y el sprint se toma por la fecha
  @CA-26.1
  Escenario: Filtrar por integrante y sprint
    Cuando consulto los registros de "Carolina" en el "Sprint 1"
    Entonces veo solo los registros "Pantalla"
    Y el total de horas de la lista es "3"

  # Caso ALTERNATIVO (CA-26.3): horas estimadas vs reales por historia
  @CA-26.1 @CA-26.3
  Escenario: Ver horas estimadas y reales por historia
    Cuando consulto los registros de la historia "HU-1"
    Entonces veo solo los registros "Validaciones, Pantalla"
    Y la historia "HU-1" tiene "6" horas estimadas y "5" horas reales

  # Caso NORMAL (CA-26.2): cada uno corrige sus registros
  @CA-26.2
  Escenario: Corregir un registro propio
    Cuando "Carolina" corrige el registro "Pantalla" a "3,5" horas
    Entonces el registro queda corregido
    Y la historia "HU-1" tiene "6" horas estimadas y "5,5" horas reales

  # Caso ALTERNATIVO (CA-26.2): el Agile Enabler corrige los de todos
  @CA-26.2
  Escenario: El Agile Enabler corrige el registro de otro
    Cuando "Juan Pablo" corrige el registro "Validaciones" a "1" horas
    Entonces el registro queda corregido

  # Caso de ERROR (CA-26.2): no se corrige el registro de otro
  @CA-26.2
  Escenario: Un integrante no corrige el registro de otro
    Cuando "Carolina" corrige el registro "Validaciones" a "1" horas
    Entonces la corrección da el error "cada integrante corrige solo sus registros (el Agile Enabler, todos)"

  # Caso LÍMITE: la corrección valida como HU-24
  @CA-26.2
  Escenario: Las horas corregidas siguen las reglas de HU-24
    Cuando "Carolina" corrige el registro "Pantalla" a "25" horas
    Entonces la corrección da el error "las horas tienen que ser mayores a 0 y hasta 24"

  # Caso de ERROR (CA-26.1): un filtro que no existe se ignora
  @CA-26.1
  Escenario: Un integrante que no existe se ignora
    Cuando consulto los registros de "Pedro" en el "Sprint 9"
    Entonces veo los 3 registros
