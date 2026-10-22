# language: es
@HU-29
Característica: Consultar defectos
  Como integrante
  Quiero ver los defectos
  Para priorizar cuáles arreglar

  Antecedentes:
    Dado el proyecto con los sprints 1 y 2 y estos defectos para consultar:
      | descripción     | severidad | detectado en | estado      |
      | Login lento     | Alta      | Sprint 1     | Abierto     |
      | Botón roto      | Crítica   | Sprint 1     | En progreso |
      | Texto cortado   | Baja      | Sprint 2     | Abierto     |
      | Error al cerrar | Alta      | Sprint 2     | Abierto     |

  # Caso NORMAL (CA-29.1): los filtros se combinan
  @CA-29.1
  Escenario: Filtrar por severidad, estado y sprint
    Cuando consulto los defectos de severidad "Alta", estado "Abierto" y sprint "Sprint 2"
    Entonces veo solo los defectos "Error al cerrar"

  # Caso ALTERNATIVO (CA-29.1 y CA-29.2): un solo filtro y la cantidad por severidad
  @CA-29.1 @CA-29.2
  Escenario: Filtrar por estado y contar por severidad
    Cuando consulto los defectos de estado "Abierto"
    Entonces veo solo los defectos "Error al cerrar, Texto cortado, Login lento"
    Y la cantidad por severidad es Crítica 0, Alta 2, Media 0 y Baja 1

  # Caso LÍMITE: nada coincide
  @CA-29.1 @CA-29.2
  Escenario: Ningún defecto cumple los filtros
    Cuando consulto los defectos de severidad "Media"
    Entonces no veo ningún defecto
    Y la cantidad por severidad es Crítica 0, Alta 0, Media 0 y Baja 0

  # Caso de ERROR: un valor que no existe se ignora
  @CA-29.1
  Escenario: Un estado que no existe se ignora
    Cuando consulto los defectos de estado "Hecho"
    Entonces veo los 4 defectos
