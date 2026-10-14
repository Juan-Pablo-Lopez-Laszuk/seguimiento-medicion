# language: es
@HU-07
Característica: Gestionar criterios de aceptación
  Como Product Builder
  Quiero cargar y marcar como cumplidos los criterios de aceptación de una historia
  Para verificar que está terminada

  Antecedentes:
    Dado el proyecto "Software Metrics" con la historia "Iniciar sesión con Google"

  # Caso NORMAL (CA-07.2 y CA-07.3): el criterio nace no cumplido y se ve en el progreso
  @CA-07.2 @CA-07.3
  Escenario: Agregar el primer criterio de la historia
    Cuando agrego a la historia el criterio "Entra con su cuenta de Google"
    Entonces el criterio "Entra con su cuenta de Google" queda no cumplido
    Y el progreso de la historia es "0/1"

  # Caso ALTERNATIVO: los criterios se muestran en el orden en que se cargaron (CA-07.3)
  @CA-07.3
  Escenario: Agregar varios criterios
    Dado que la historia tiene el criterio "Entra con su cuenta de Google"
    Cuando agrego a la historia el criterio "Muestra su nombre"
    Y agrego a la historia el criterio "Puede cerrar sesión"
    Entonces los criterios de la historia, en orden, son "Entra con su cuenta de Google; Muestra su nombre; Puede cerrar sesión"
    Y el progreso de la historia es "0/3"

  # Caso ALTERNATIVO: marcar y desmarcar de forma explícita (CA-07.2 y CA-07.3)
  @CA-07.2 @CA-07.3
  Escenario: Marcar un criterio como cumplido y volver a desmarcarlo
    Dado que la historia tiene 3 criterios, 0 de ellos cumplidos
    Cuando marco como cumplido el criterio número 1
    Entonces el progreso de la historia es "1/3"
    Cuando marco como no cumplido el criterio número 1
    Entonces el progreso de la historia es "0/3"

  # Caso LÍMITE: marcar dos veces lo mismo no lo deja al revés (RN3)
  @CA-07.2
  Escenario: Marcar dos veces un criterio como cumplido
    Dado que la historia tiene 2 criterios, 0 de ellos cumplidos
    Cuando marco como cumplido el criterio número 2
    Y marco como cumplido el criterio número 2
    Entonces el criterio número 2 queda cumplido
    Y el progreso de la historia es "1/2"

  # Caso ALTERNATIVO: cada historia cuenta solo sus criterios (RN8)
  @CA-07.3
  Escenario: Cada historia tiene sus propios criterios
    Dado que la historia tiene 2 criterios, 1 de ellos cumplidos
    Y la historia "Ver el tablero" con 1 criterio, 0 de ellos cumplidos
    Entonces el progreso de la historia es "1/2"
    Y el progreso de la historia "Ver el tablero" es "0/1"

  # Caso LÍMITE: se necesita al menos un criterio para entrar a un sprint (CA-07.1)
  @CA-07.1
  Esquema del escenario: Una historia necesita criterios para entrar a un sprint
    Dado que la historia tiene <total> criterios, <cumplidos> de ellos cumplidos
    Entonces la historia <entrada> entrar a un sprint por sus criterios

    Ejemplos:
      | total | cumplidos | entrada   |
      | 0     | 0         | no puede  |
      | 1     | 0         | puede     |
      | 3     | 3         | puede     |

  # Caso LÍMITE: completa quiere decir con criterios y todos cumplidos; 0/0 no cuenta (RN6)
  @CA-07.3
  Esquema del escenario: Cuándo una historia está completa
    Dado que la historia tiene <total> criterios, <cumplidos> de ellos cumplidos
    Entonces la historia <completa> completa

    Ejemplos:
      | total | cumplidos | completa |
      | 0     | 0         | no está  |
      | 2     | 0         | no está  |
      | 2     | 1         | no está  |
      | 2     | 2         | está     |

  # Caso LÍMITE: la descripción tiene hasta 300 caracteres (RN1)
  @CA-07.2
  Esquema del escenario: Largo de la descripción del criterio
    Cuando agrego a la historia un criterio de <largo> caracteres
    Entonces el criterio da como resultado "<resultado>"

    Ejemplos:
      | largo | resultado                                                                    |
      | 300   | agregado                                                                     |
      | 301   | descripcion: la descripción del criterio no puede superar los 300 caracteres |

  # Caso de ERROR: descripción vacía
  @CA-07.2
  Escenario: La descripción del criterio es obligatoria
    Cuando agrego a la historia el criterio "   "
    Entonces el criterio da como resultado "descripcion: la descripción del criterio es obligatoria"
    Y el progreso de la historia es "0/0"

  # Caso de ERROR: una historia Hecha no admite cambios en sus criterios (RN7)
  @CA-07.2
  Escenario: Una historia Hecha no admite cambios en sus criterios
    Dado que la historia "Registrar horas" está Hecha y tiene 2 criterios cumplidos
    Cuando agrego a la historia "Registrar horas" el criterio "Nuevo criterio"
    Entonces el criterio da como resultado "la historia está Hecha: no se pueden cambiar sus criterios"
    Cuando marco como no cumplido el criterio número 1 de la historia "Registrar horas"
    Entonces el criterio da como resultado "la historia está Hecha: no se pueden cambiar sus criterios"
    Y el progreso de la historia "Registrar horas" es "2/2"
