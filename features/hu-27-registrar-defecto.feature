# language: es
@HU-27
Característica: Registrar defecto
  Como integrante
  Quiero registrar un defecto
  Para darle seguimiento

  Antecedentes:
    Dado el proyecto "Software Metrics" con el Sprint 1 y la historia "Iniciar sesión con Google"

  # Caso NORMAL (CA-27.1, CA-27.2, CA-27.3)
  @CA-27.1 @CA-27.2 @CA-27.3
  Escenario: Registrar un defecto de una historia
    Cuando registro el defecto "El login falla con Google" de severidad "Alta" en el Sprint 1 para la historia
    Entonces el defecto queda "Abierto" con severidad "Alta"
    Y el proyecto tiene 1 defecto

  # Caso ALTERNATIVO: la historia es opcional
  @CA-27.3
  Escenario: Registrar un defecto sin historia
    Cuando registro el defecto "La página tarda en cargar" de severidad "Baja" en el Sprint 1
    Entonces el defecto queda "Abierto" con severidad "Baja"

  # Caso de ERROR (CA-27.1)
  @CA-27.1
  Esquema del escenario: Datos inválidos de un defecto
    Cuando registro el defecto "<descripcion>" de severidad "<severidad>" en el Sprint 1
    Entonces el defecto se rechaza con el error "<error>"
    Y el proyecto tiene 0 defectos

    Ejemplos:
      | descripcion               | severidad | error                                                  |
      |                           | Alta      | la descripción del defecto es obligatoria              |
      | El login falla con Google | Urgente   | la severidad tiene que ser Crítica, Alta, Media o Baja |

  # Caso de ERROR (CA-27.3): sprint e historia del proyecto
  @CA-27.3
  Escenario: El sprint y la historia tienen que ser del proyecto
    Dado otro proyecto "Otro" con su Sprint 1 y la historia "Ajena"
    Cuando registro un defecto con el sprint y la historia del otro proyecto
    Entonces el defecto se rechaza con el error "el sprint de detección tiene que ser un sprint del proyecto"
    Y el defecto se rechaza con el error "la historia tiene que ser del proyecto"
