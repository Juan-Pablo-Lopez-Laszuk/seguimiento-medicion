# language: es
@HU-03
Característica: Registrar integrantes
  Como Agile Enabler
  Quiero registrar integrantes con nombre, email y rol
  Para asignarles trabajo y registrar su esfuerzo

  Antecedentes:
    Dado el proyecto "Software Metrics" para registrar integrantes

  # Caso NORMAL (CA-03.1 y CA-03.2)
  @CA-03.1 @CA-03.2
  Escenario: Registrar al Agile Enabler del proyecto
    Cuando registro a "Juan Pablo" con el email " JuanPablo@Mail.com " como "Agile Enabler"
    Entonces queda registrado "Juan Pablo" con el email "juanpablo@mail.com" como "Agile Enabler"
    Y el integrante está activo

  # Caso ALTERNATIVO: puede haber varios Product Builder (CA-03.2)
  @CA-03.2
  Escenario: Registrar varios Product Builder
    Dado que ya está registrado "Mariano" con el email "mariano@mail.com" como "Product Builder"
    Cuando registro a "Carolina" con el email "carolina@mail.com" como "Product Builder"
    Entonces el proyecto tiene 2 integrantes activos

  # Caso ALTERNATIVO: el mismo email puede estar en otro proyecto (CA-03.1)
  @CA-03.1
  Escenario: El mismo email en otro proyecto
    Dado que ya está registrado "Mariano" con el email "mariano@mail.com" como "Product Builder"
    Y el proyecto "Otro proyecto" para registrar integrantes
    Cuando registro en "Otro proyecto" a "Mariano" con el email "mariano@mail.com" como "Product Builder"
    Entonces queda registrado "Mariano" con el email "mariano@mail.com" como "Product Builder"

  # Caso LÍMITE: formato del email (CA-03.1)
  @CA-03.1
  Esquema del escenario: Formato del email
    Cuando registro a "Carolina" con el email "<email>" como "Product Builder"
    Entonces el registro da como resultado "<resultado>"

    Ejemplos:
      | email                | resultado                                  |
      | carolina@mail.com    | registrado                                 |
      | carolina@mail        | email: el email no tiene un formato válido |
      | carolina.mail.com    | email: el email no tiene un formato válido |
      | Caro <caro@mail.com> | email: el email no tiene un formato válido |
      |                      | email: el email es obligatorio             |

  # Caso de ERROR: email repetido, también contra uno dado de baja (CA-03.1)
  @CA-03.1
  Escenario: No se repite el email dentro del proyecto
    Dado que ya está registrado "Mariano" con el email "mariano@mail.com" como "Product Builder"
    Y "Mariano" está dado de baja
    Cuando registro a "Mariano P." con el email "MARIANO@mail.com" como "Product Builder"
    Entonces el registro da como resultado "email: ya hay un integrante con ese email en el proyecto"

  # Caso de ERROR: un solo Agile Enabler activo (CA-03.2)
  @CA-03.2
  Escenario: No puede haber dos Agile Enabler activos
    Dado que ya está registrado "Juan Pablo" con el email "jp@mail.com" como "Agile Enabler"
    Cuando registro a "Carolina" con el email "carolina@mail.com" como "Agile Enabler"
    Entonces el registro da como resultado "rol: el proyecto ya tiene un Agile Enabler activo"

  # Caso ALTERNATIVO: si el Agile Enabler se da de baja, se puede registrar otro (CA-03.2)
  @CA-03.2
  Escenario: Nuevo Agile Enabler después de una baja
    Dado que ya está registrado "Juan Pablo" con el email "jp@mail.com" como "Agile Enabler"
    Y "Juan Pablo" está dado de baja
    Cuando registro a "Carolina" con el email "carolina@mail.com" como "Agile Enabler"
    Entonces queda registrado "Carolina" con el email "carolina@mail.com" como "Agile Enabler"

  # Caso de ERROR: rol desconocido y nombre vacío, informados juntos (CA-03.2)
  @CA-03.2
  Escenario: Se indica cada campo que falló
    Cuando registro a "   " con el email "ana@mail.com" como "Scrum Master"
    Entonces el registro da como resultado "nombre: el nombre es obligatorio"
    Y el registro da como resultado "rol: el rol tiene que ser Agile Enabler o Product Builder"
    Y el proyecto tiene 0 integrantes activos

  # Caso NORMAL de la baja (CA-03.3)
  @CA-03.3
  Escenario: Dar de baja a un integrante
    Dado que ya está registrado "Mariano" con el email "mariano@mail.com" como "Product Builder"
    Cuando doy de baja a "Mariano"
    Entonces "Mariano" sigue en el proyecto pero dado de baja
    Y el proyecto tiene 0 integrantes activos

  # Caso de ERROR de la baja (CA-03.3)
  @CA-03.3
  Escenario: No se da de baja dos veces
    Dado que ya está registrado "Mariano" con el email "mariano@mail.com" como "Product Builder"
    Y "Mariano" está dado de baja
    Cuando doy de baja a "Mariano"
    Entonces la baja da el error "el integrante ya está dado de baja"
