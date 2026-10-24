# language: es
@HU-09
Característica: Priorizar el backlog
  Como equipo
  Queremos ordenar el backlog por prioridad y, dentro de cada prioridad, a mano
  Para decidir qué entra en el próximo sprint

  Antecedentes:
    Dado el proyecto "Software Metrics" con estas historias para priorizar:
      | título   | prioridad | estado    |
      | Login    | Baja      | Pendiente |
      | Registro | Alta      | Pendiente |
      | Perfil   | Media     | Pendiente |
      | Reportes | Alta      | Pendiente |

  # Caso NORMAL (CA-09.1, RN1): primero la prioridad; dentro de cada una, por número de historia
  @CA-09.1 @RN1
  Escenario: El backlog se ordena por prioridad
    Entonces el backlog queda ordenado así:
      | título   |
      | Registro |
      | Reportes |
      | Perfil   |
      | Login    |

  # Caso NORMAL (CA-09.1, RN2): subir una historia un lugar dentro de su prioridad
  @CA-09.1 @RN2
  Escenario: Subir una historia
    Cuando subo la historia "Reportes"
    Entonces el backlog queda ordenado así:
      | título   |
      | Reportes |
      | Registro |
      | Perfil   |
      | Login    |

  # Caso ALTERNATIVO (CA-09.1, RN2): bajar una historia un lugar dentro de su prioridad
  @CA-09.1 @RN2
  Escenario: Bajar una historia
    Cuando bajo la historia "Registro"
    Entonces el backlog queda ordenado así:
      | título   |
      | Reportes |
      | Registro |
      | Perfil   |
      | Login    |

  # Caso ALTERNATIVO (CA-09.1, RN2): se mueve de a un lugar, así que para llegar más arriba hay que subir varias veces
  @CA-09.1 @RN2
  Escenario: Subir varias veces
    Dado que se agrega al backlog la historia "Cuentas" con prioridad "Alta"
    Cuando subo la historia "Cuentas"
    Y subo la historia "Cuentas"
    Entonces el backlog queda ordenado así:
      | título   |
      | Cuentas  |
      | Registro |
      | Reportes |
      | Perfil   |
      | Login    |

  # Caso LÍMITE y de ERROR (RN3): la primera de su prioridad no sube y la última no baja; no se cruza a otra prioridad
  @CA-09.1 @RN3
  Esquema del escenario: No se sube la primera ni se baja la última de su prioridad
    Cuando intento <movimiento> la historia "<historia>"
    Entonces el movimiento falla con "<error>"
    Y el backlog sigue ordenado así:
      | título   |
      | Registro |
      | Reportes |
      | Perfil   |
      | Login    |

    Ejemplos:
      | movimiento | historia | error                                  |
      | subir      | Registro | HU-2 ya es la primera de su prioridad  |
      | bajar      | Reportes | HU-4 ya es la última de su prioridad   |
      | subir      | Perfil   | HU-3 ya es la primera de su prioridad  |
      | bajar      | Perfil   | HU-3 ya es la última de su prioridad   |
      | subir      | Login    | HU-1 ya es la primera de su prioridad  |

  # Caso NORMAL (CA-09.2, RN5): el orden manual se guarda
  @CA-09.2 @RN5
  Escenario: El orden manual se guarda
    Cuando subo la historia "Reportes"
    Entonces al volver a cargar el backlog desde lo guardado queda ordenado así:
      | título   |
      | Reportes |
      | Registro |
      | Perfil   |
      | Login    |

  # Caso ALTERNATIVO (CA-09.2, RN6): una historia nueva va al final de su prioridad, también después de reordenar
  @CA-09.2 @RN6
  Escenario: Una historia nueva va al final de su prioridad
    Dado que ya subí la historia "Reportes"
    Cuando se agrega al backlog la historia "Cuentas" con prioridad "Alta"
    Entonces el backlog queda ordenado así:
      | título   |
      | Reportes |
      | Registro |
      | Cuentas  |
      | Perfil   |
      | Login    |

  # Caso ALTERNATIVO (RN7): al cambiar la prioridad la historia conserva su orden
  @CA-09.1 @RN7
  Escenario: Cambiar la prioridad conserva el orden de la historia
    Cuando cambio la prioridad de "Login" a "Alta"
    Entonces el backlog queda ordenado así:
      | título   |
      | Login    |
      | Registro |
      | Reportes |
      | Perfil   |

  # Caso ALTERNATIVO (RN8): una historia asignada a un sprint sigue en el backlog y se puede reordenar
  @CA-09.1 @RN8
  Escenario: Una historia de un sprint se reordena igual
    Dado que "Registro" está en el estado "En sprint"
    Cuando bajo la historia "Registro"
    Entonces el backlog queda ordenado así:
      | título   |
      | Reportes |
      | Registro |
      | Perfil   |
      | Login    |

  # Caso NORMAL (CA-09.3, RN4): las Hechas se muestran aparte, la más reciente primero
  @CA-09.3 @RN4
  Escenario: Las historias Hechas se muestran aparte
    Dado que "Login" está Hecha desde el "2026-10-12"
    Y que "Perfil" está Hecha desde el "2026-10-14"
    Entonces el backlog queda ordenado así:
      | título   |
      | Registro |
      | Reportes |
    Y las historias Hechas son:
      | título |
      | Perfil |
      | Login  |

  # Caso de ERROR (RN4): una Hecha no se reordena
  @CA-09.3 @RN4
  Escenario: Una historia Hecha no se reordena
    Dado que "Login" está Hecha desde el "2026-10-12"
    Cuando intento subir la historia "Login"
    Entonces el movimiento falla con "una historia Hecha no se reordena"

  # Caso de ERROR: una historia que no existe en el proyecto
  @CA-09.1
  Escenario: Mover una historia que no existe
    Cuando intento subir la historia número 99
    Entonces el movimiento falla con "no existe la historia HU-99"

  # Caso ALTERNATIVO (RN9): la página del sprint muestra las candidatas en el mismo orden que el backlog
  @CA-09.1 @RN9
  Escenario: Las candidatas de un sprint siguen el orden del backlog
    Dado que existe el Sprint 1 del proyecto
    Cuando subo la historia "Reportes"
    Entonces las candidatas del Sprint 1 están en este orden:
      | título   |
      | Reportes |
      | Registro |
      | Perfil   |
      | Login    |
