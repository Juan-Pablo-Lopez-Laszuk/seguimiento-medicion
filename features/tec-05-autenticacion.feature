# language: es
@TEC-05
Característica: Autenticación de usuarios con Supabase Auth
  Como integrante del equipo
  Quiero registrarme e iniciar sesión
  Para que el sistema sepa quién soy y a qué integrante de cada proyecto corresponde mi cuenta

  # Caso NORMAL (CA-05.1): me registro y la cuenta queda creada
  @CA-05.1
  Escenario: Registrarse con datos válidos
    Dado que el servicio de login está disponible
    Cuando me registro con el email "ana@mail.com" y la contraseña "clave-segura-1"
    Entonces la cuenta "ana@mail.com" existe en el servicio de login
    Y la respuesta del registro es el mensaje "Cuenta creada. Ya podés iniciar sesión."

  # Caso ALTERNATIVO (RN5): Supabase pide confirmar el email antes de dejar entrar
  @CA-05.1 @RN5
  Escenario: Registrarse cuando el servicio pide confirmar el email
    Dado que el servicio de login pide confirmar el email
    Cuando me registro con el email "ana@mail.com" y la contraseña "clave-segura-1"
    Entonces la respuesta del registro es el mensaje "Te enviamos un email para confirmar la cuenta. Confirmala y después iniciá sesión."

  # Caso LÍMITE y de ERROR (RN5): largo de la contraseña, repetición y email
  @CA-05.1 @RN5
  Esquema del escenario: Datos de registro inválidos
    Dado que el servicio de login está disponible
    Cuando me registro con el email "<email>", la contraseña "<clave>" y la repetición "<repeticion>"
    Entonces el registro falla en el campo "<campo>" con "<mensaje>"
    Y no se crea ninguna cuenta

    Ejemplos:
      | email         | clave                | repeticion           | campo      | mensaje                                                         |
      | ana@mail.com  | corta12              | corta12              | clave      | la contraseña tiene que tener entre 8 y 72 caracteres          |
      | ana@mail.com  | clave-segura-1       | clave-segura-2       | repeticion | las contraseñas no coinciden                                    |
      | ana@mail      | clave-segura-1       | clave-segura-1       | email      | el email no tiene un formato válido                             |
      |               | clave-segura-1       | clave-segura-1       | email      | el email es obligatorio                                         |

  # Caso LÍMITE (RN5): 8 y 72 caracteres son válidos
  @CA-05.1 @RN5
  Esquema del escenario: Largos válidos de contraseña
    Dado que el servicio de login está disponible
    Cuando me registro con el email "ana@mail.com" y una contraseña de <largo> caracteres
    Entonces la cuenta "ana@mail.com" existe en el servicio de login

    Ejemplos:
      | largo |
      | 8     |
      | 72    |

  # Caso de ERROR (RN5): el email ya tiene cuenta
  @CA-05.1 @RN5
  Escenario: Registrarse con un email que ya tiene cuenta
    Dado que el servicio de login está disponible
    Y que ya existe la cuenta "ana@mail.com" con la contraseña "clave-segura-1"
    Cuando me registro con el email "ana@mail.com" y la contraseña "otra-clave-9"
    Entonces el registro falla en el campo "email" con "ya hay una cuenta con ese email"

  # Caso NORMAL (CA-05.1): ingreso con las credenciales correctas
  @CA-05.1
  Escenario: Iniciar sesión con las credenciales correctas
    Dado que el servicio de login está disponible
    Y que ya existe la cuenta "ana@mail.com" con la contraseña "clave-segura-1"
    Cuando inicio sesión con el email "Ana@Mail.com" y la contraseña "clave-segura-1"
    Entonces el ingreso es correcto y se guarda la cookie de sesión
    Y la cookie de sesión es "HttpOnly" y "SameSite=Lax"

  # Caso de ERROR (RN6): no se dice cuál de los dos datos falló
  @CA-05.1 @RN6
  Esquema del escenario: Iniciar sesión con credenciales incorrectas
    Dado que el servicio de login está disponible
    Y que ya existe la cuenta "ana@mail.com" con la contraseña "clave-segura-1"
    Cuando inicio sesión con el email "<email>" y la contraseña "<clave>"
    Entonces el ingreso falla con el código 401 y el mensaje "El email o la contraseña no son correctos."
    Y no se guarda ninguna cookie de sesión

    Ejemplos:
      | email            | clave          |
      | ana@mail.com     | otra-clave-9   |
      | nadie@mail.com   | clave-segura-1 |

  # Caso de ERROR (RN12): el servicio de login no responde
  @CA-05.1 @RN12
  Escenario: El servicio de login no responde
    Dado que el servicio de login no responde
    Cuando inicio sesión con el email "ana@mail.com" y la contraseña "clave-segura-1"
    Entonces el ingreso falla con el código 503 y el mensaje "No pudimos conectar con el servicio de inicio de sesión. Probá de nuevo en unos minutos."

  # Caso NORMAL (CA-05.2): con una sesión válida se llega a la página
  @CA-05.2
  Escenario: Con una sesión válida se accede a la aplicación
    Dado que el login está activado
    Y que tengo una sesión válida de "ana@mail.com"
    Cuando pido la página "/proyectos"
    Entonces la respuesta tiene el código 200

  # Caso de ERROR (CA-05.2, RN3): sin sesión o con un token que no vale, 401
  @CA-05.2 @RN3
  Esquema del escenario: Un pedido sin una sesión válida se rechaza
    Dado que el login está activado
    Y que mi cookie de sesión es <cookie>
    Cuando pido la página "/proyectos"
    Entonces la respuesta tiene el código 401
    Y la respuesta muestra el mensaje "<mensaje>"
    Y la respuesta no muestra la lista de proyectos

    Ejemplos:
      | cookie                                         | mensaje                                   |
      | ausente                                        | Iniciá sesión para continuar.             |
      | un texto que no es un token                    | Iniciá sesión para continuar.             |
      | un token vencido                               | Tu sesión venció, iniciá sesión de nuevo. |
      | un token firmado con otra clave                | Iniciá sesión para continuar.             |
      | un token sin firma ("alg: none")               | Iniciá sesión para continuar.             |
      | un token con algoritmo HS256                   | Iniciá sesión para continuar.             |
      | un token de otro emisor                        | Iniciá sesión para continuar.             |
      | un token con otra audiencia                    | Iniciá sesión para continuar.             |
      | un token cuyo "sub" no es un UUID              | Iniciá sesión para continuar.             |

  # Caso LÍMITE (RN3): un pedido que modifica datos también se rechaza, y no se ejecuta
  @CA-05.2 @RN3
  Escenario: Un POST sin sesión no ejecuta nada
    Dado que el login está activado
    Cuando creo el proyecto "Sin sesión" sin haber iniciado sesión
    Entonces la respuesta tiene el código 401
    Y no se creó ningún proyecto

  # Caso ALTERNATIVO (RN1): lo que queda abierto con el login activado
  @CA-05.2 @RN1
  Esquema del escenario: Rutas abiertas con el login activado
    Dado que el login está activado
    Cuando pido la página "<ruta>" sin haber iniciado sesión
    Entonces la respuesta tiene el código 200

    Ejemplos:
      | ruta      |
      | /health   |
      | /login    |
      | /registro |

  # Caso ALTERNATIVO (RN1): sin configuración no hay login (previews, tests y CI)
  @CA-05.2 @RN1
  Escenario: Sin configuración la aplicación no pide login
    Dado que el login no está configurado
    Cuando pido la página "/proyectos" sin haber iniciado sesión
    Entonces la respuesta tiene el código 200

  # Caso de ERROR (RN1): configuración a medias
  @CA-05.2 @RN1
  Esquema del escenario: Configuración incompleta
    Dado que solo está definida la variable <definida>
    Cuando arranca la aplicación
    Entonces la aplicación no arranca y dice "falta <faltante>"

    Ejemplos:
      | definida          | faltante          |
      | SUPABASE_URL      | SUPABASE_ANON_KEY |
      | SUPABASE_ANON_KEY | SUPABASE_URL      |

  # Caso NORMAL (RN9): después de ingresar se vuelve a donde se estaba
  @CA-05.1 @RN9
  Esquema del escenario: Volver a donde se estaba después de ingresar
    Dado que el servicio de login está disponible
    Y que ya existe la cuenta "ana@mail.com" con la contraseña "clave-segura-1"
    Cuando inicio sesión con la contraseña "clave-segura-1" y siguiente "<siguiente>"
    Entonces el ingreso redirige a "<destino>"

    Ejemplos:
      | siguiente            | destino              |
      | /proyectos/1/backlog | /proyectos/1/backlog |
      |                      | /                    |
      | https://malo.com     | /                    |
      | //malo.com           | /                    |
      | /\malo.com           | /                    |

  # Caso NORMAL (RN10): salir borra la cookie
  @CA-05.1 @RN10
  Escenario: Cerrar sesión
    Dado que el login está activado
    Y que tengo una sesión válida de "ana@mail.com"
    Cuando cierro sesión
    Entonces la cookie de sesión se borra
    Y la respuesta lleva a "/login" con el mensaje "Cerraste sesión."

  # Caso NORMAL (CA-05.3, RN7): el usuario se vincula con los integrantes que tienen su email
  @CA-05.3 @RN7
  Escenario: Al ingresar la cuenta se vincula con los integrantes de todos los proyectos
    Dado que el servicio de login está disponible
    Y que ya existe la cuenta "ana@mail.com" con la contraseña "clave-segura-1"
    Y que "Ana" con el email "ana@mail.com" es integrante de los proyectos "Alfa" y "Beta"
    Y que "Beto" con el email "beto@mail.com" es integrante del proyecto "Alfa"
    Cuando inicio sesión con el email "ana@mail.com" y la contraseña "clave-segura-1"
    Entonces "Ana" queda vinculada con la cuenta "ana@mail.com" en los proyectos "Alfa" y "Beta"
    Y "Beto" no queda vinculado con ninguna cuenta

  # Caso LÍMITE (RN7): un integrante ya vinculado con otra cuenta no se pisa
  @CA-05.3 @RN7
  Escenario: Un integrante ya vinculado con otra cuenta no se pisa
    Dado que el servicio de login está disponible
    Y que ya existe la cuenta "ana@mail.com" con la contraseña "clave-segura-1"
    Y que "Ana" con el email "ana@mail.com" es integrante del proyecto "Alfa" y ya está vinculada con otra cuenta
    Cuando inicio sesión con el email "ana@mail.com" y la contraseña "clave-segura-1"
    Entonces "Ana" sigue vinculada con la otra cuenta

  # Caso NORMAL (CA-05.3, RN8): saber quién es el usuario en un proyecto
  @CA-05.3 @RN8
  Escenario: Saber a qué integrante corresponde el usuario en un proyecto
    Dado que "Ana" con el email "ana@mail.com" es integrante del proyecto "Alfa"
    Y que inicié sesión como "ana@mail.com"
    Cuando consulto quién soy en el proyecto "Alfa"
    Entonces soy el integrante "Ana"

  # Caso ALTERNATIVO (RN8): me agregaron al proyecto después de ingresar
  @CA-05.3 @RN8
  Escenario: Me agregan a un proyecto después de haber ingresado
    Dado que hay un proyecto "Alfa" sin integrantes
    Y que inicié sesión como "ana@mail.com"
    Y que después "Ana" con el email "ana@mail.com" se registra como integrante del proyecto "Alfa"
    Cuando consulto quién soy en el proyecto "Alfa"
    Entonces soy el integrante "Ana"
    Y "Ana" queda vinculada con la cuenta "ana@mail.com" en el proyecto "Alfa"

  # Caso de ERROR (RN8): no soy integrante de ese proyecto
  @CA-05.3 @RN8
  Escenario: Un usuario que no es integrante del proyecto
    Dado que "Beto" con el email "beto@mail.com" es integrante del proyecto "Alfa"
    Y que inicié sesión como "ana@mail.com"
    Cuando consulto quién soy en el proyecto "Alfa"
    Entonces el resultado es que no soy integrante del proyecto
