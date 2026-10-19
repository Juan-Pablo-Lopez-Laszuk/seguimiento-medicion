# language: es
@HU-28
Característica: Gestionar el estado de un defecto
  Como integrante
  Quiero cambiar el estado de un defecto
  Para registrar cuándo se resuelve

  Antecedentes:
    Dado el proyecto con los sprints 1 y 2 y un defecto Abierto detectado en el Sprint 1

  # Caso NORMAL (CA-28.1 y CA-28.2)
  @CA-28.1 @CA-28.2
  Escenario: Llevar un defecto hasta Cerrado
    Cuando paso el defecto a "En progreso"
    Y lo resuelvo en el Sprint 2
    Y paso el defecto a "Cerrado"
    Entonces el defecto queda "Cerrado" con sprint de resolución "Sprint 2"

  # Caso LÍMITE (CA-28.2): el mismo sprint de detección vale
  @CA-28.2
  Escenario: Resolver en el mismo sprint en que se detectó
    Dado que el defecto está "En progreso"
    Cuando lo resuelvo en el Sprint 1
    Entonces el defecto queda "Resuelto" con sprint de resolución "Sprint 1"

  # Caso ALTERNATIVO (CA-28.3): reabrir borra el sprint de resolución
  @CA-28.3
  Escenario: Reabrir un defecto resuelto
    Dado que el defecto está "En progreso"
    Y lo resuelvo en el Sprint 2
    Cuando paso el defecto a "Reabierto"
    Entonces el defecto queda "Reabierto" sin sprint de resolución

  # Caso de ERROR (CA-28.1): transiciones que no valen
  @CA-28.1
  Esquema del escenario: Transiciones no permitidas
    Cuando paso el defecto a "<estado>"
    Entonces el cambio se rechaza con el error "un defecto en estado Abierto no puede pasar a <estado>"
    Y el defecto queda "Abierto" sin sprint de resolución

    Ejemplos:
      | estado    |
      | Resuelto  |
      | Cerrado   |
      | Reabierto |
