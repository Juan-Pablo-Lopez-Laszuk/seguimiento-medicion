# language: es
@HU-33
Característica: Horas estimadas, reales y desviación
  Como Agile Enabler
  Quiero comparar horas estimadas y reales
  Para estimar mejor

  # Caso NORMAL (CA-33.1): se tardó más de lo estimado
  @CA-33.1
  Escenario: Una historia que llevó más horas de las estimadas
    Dado una historia con tareas estimadas en 4 h y 6 h
    Cuando se registran 12 h de trabajo en la historia
    Entonces la desviación de la historia es "+2 h" y "+20,0 %"

  # Caso ALTERNATIVO (CA-33.1): se tardó menos
  @CA-33.1
  Escenario: Una historia que llevó menos horas de las estimadas
    Dado una historia con tareas estimadas en 4 h y 6 h
    Cuando se registran 8 h de trabajo en la historia
    Entonces la desviación de la historia es "-2 h" y "-20,0 %"

  # Caso LÍMITE (CA-33.2): sin horas estimadas no hay porcentaje
  @CA-33.2
  Escenario: Una historia sin horas estimadas
    Dado una historia sin tareas estimadas
    Cuando se registran 3 h de trabajo en la historia
    Entonces la desviación de la historia es "+3 h" y "no aplicable"

  # Caso NORMAL (CA-33.3): el sprint y el proyecto suman sus historias
  @CA-33.3
  Escenario: Desviación por sprint y por proyecto
    Dado en el Sprint 1 una historia estimada en 10 h con 12 h trabajadas
    Y en el Sprint 1 una historia estimada en 5 h con 4 h trabajadas
    Y sin sprint una historia estimada en 5 h con 5 h trabajadas
    Entonces la desviación del Sprint 1 es "+1 h" y "+6,7 %"
    Y la desviación del proyecto es "+1 h" y "+5,0 %"
