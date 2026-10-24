# Trazabilidad: requerimientos de la consigna → escenarios y tests

Matriz de **TEC-06** (#69, CA-06.3). Para cada requerimiento mínimo de la [consigna](plan/consigna.pdf) dice qué historia lo
cubre y qué **escenario BDD** y qué **tests** lo prueban. Es la mitad de la cadena de trazabilidad que pide la consigna:

**Historia de Usuario → Especificación SDD → Criterios de Aceptación → Escenarios BDD → Tests → Código Go**

La otra mitad, de cada historia hacia su código, está en la sección **Trazabilidad** de su especificación en
[`specs/`](../specs/) (con la tabla *criterio → escenarios → tests* y el código de cada capa). La matriz parte de lo que
pide la consigna y lleva hasta esos mismos escenarios y tests.

## Cómo leerla

| Columna | Qué dice |
|---|---|
| Requerimiento | El requerimiento de la consigna, con su bloque (`Proyectos · Crear proyectos`) |
| Historia | La historia de usuario que lo cubre; su especificación es `specs/<historia>-….md` |
| Escenarios BDD | El archivo `.feature` de `features/` y, entre «comillas», los escenarios que lo prueban |
| Tests | Las funciones de test de Go que lo prueban (con `*` al final son todas las que empiezan así) |
| Estado | **Cubierto**, o **Pendiente** con la historia que lo va a cubrir |

## Cómo se mantiene y por qué no miente

Esta matriz la comprueba `features/trazabilidad_test.go`, que corre con `go test ./...` y en el CI. Falla si:

- falta una fila para algún requerimiento de la consigna (la lista está en el propio test) o hay una repetida;
- una fila marcada **Cubierto** no nombra ningún escenario ni ningún test;
- una fila **Pendiente** no dice qué historia la va a cubrir;
- se nombra un archivo `.feature`, un escenario o un test que no existe (por ejemplo, porque se renombró).

Cuando una historia pendiente se mergea, se pasa su fila a **Cubierto** con sus escenarios y tests. Un requerimiento nuevo se
suma a la lista del test y a la matriz.

**Estado al 22/10/2026: 50 de 59 requerimientos cubiertos.** Los 9 pendientes son de Sprint 3 y 4: Planning Poker
(HU-17 a HU-22), los gráficos del dashboard (HU-36), el reporte de un proyecto (HU-38) y la exportación a PDF (HU-39).

> Lo que no es un requerimiento de la consigna (el login de TEC-05, el deploy de TEC-04, el umbral de cobertura de TEC-06
> y el diseño base de TEC-07) se traza en su propia especificación.

## 1. Gestión de proyectos

| Requerimiento | Historia | Escenarios BDD | Tests | Estado |
|---|---|---|---|---|
| Proyectos · Crear proyectos | HU-01 | `hu-01-crear-proyecto.feature` «Crear un proyecto con datos válidos» · «No se permite un nombre repetido» | `TestNuevo_LargoDelNombre`, `TestCrear_NombreRepetido_*`, `TestCrearProyecto_DatosValidos_*` | Cubierto |
| Proyectos · Modificar proyectos | HU-02 | `hu-02-modificar-proyecto.feature` «Modificar la descripción» · «Un proyecto Finalizado no se modifica» | `TestModificar_CambiarDescripcion`, `TestModificar_ValidacionesDeLaCreacion`, `TestModificarProyecto_DatosValidos_QuedaGuardado` | Cubierto |
| Proyectos · Registrar integrantes | HU-03 | `hu-03-registrar-integrantes.feature` «Registrar al Agile Enabler del proyecto» · «Registrar varios Product Builder» · «Dar de baja a un integrante» | `TestNuevo_Rol`, `TestNuevo_UnSoloAgileEnablerActivo`, `TestNuevo_EmailRepetido`, `TestRegistrarIntegrante_DatosValidos_*` | Cubierto |
| Proyectos · Registrar fecha de inicio y finalización | HU-01 | `hu-01-crear-proyecto.feature` «Fecha de finalización en el borde» | `TestNuevo_FechasObligatorias`, `TestNuevo_FechaDeFinPosteriorALaDeInicio`, `TestNuevo_LasFechasSeTomanSinHora` | Cubierto |
| Proyectos · Consultar el estado de un proyecto | HU-04 | `hu-04-estado-proyecto.feature` «El estado depende de la fecha y los sprints» · «Proyecto en curso con su sprint activo» | `TestActivo`, `TestProximo`, `TestTerminado`, `TestConsultarEstado_ConSprintActivo` | Cubierto |

## 2. Product Backlog

| Requerimiento | Historia | Escenarios BDD | Tests | Estado |
|---|---|---|---|---|
| Product Backlog · Identificador | HU-05, HU-06 | `hu-05-crear-item-backlog.feature` «La siguiente historia lleva el número que sigue» · «Cada proyecto numera sus propias historias»; `hu-06-editar-eliminar-item.feature` «Editar los datos no cambia el identificador» | `TestNueva_ElNumeroSigueAlMasAlto`, `TestNueva_NoReutilizaNumerosQueFaltan`, `TestHistorias_LaBaseRechazaUnNumeroRepetidoEnElProyecto`, `TestEditar_CambiaLosDatosYConservaElIdentificador` | Cubierto |
| Product Backlog · Título | HU-05 | `hu-05-crear-item-backlog.feature` «Largo del título» · «El título es obligatorio» | `TestNueva_LargoDelTitulo`, `TestNueva_QuitaLosEspaciosAlrededorDelTitulo`, `TestNueva_TituloVacioOSoloEspacios_*` | Cubierto |
| Product Backlog · Descripción | HU-05, HU-06 | `hu-05-crear-item-backlog.feature` «Crear la primera historia del backlog»; `hu-06-editar-eliminar-item.feature` «Largo del título y de la descripción» | `TestNueva_LargoDeLaDescripcion`, `TestNueva_Descripcion_*`, `TestEditar_LargoDeLaDescripcion` | Cubierto |
| Product Backlog · Prioridad | HU-05, HU-09 | `hu-05-crear-item-backlog.feature` «Prioridad de la historia»; `hu-09-priorizar-backlog.feature` «El backlog se ordena por prioridad» · «Subir una historia» · «El orden manual se guarda» | `TestNueva_PrioridadesValidas`, `TestNueva_PrioridadInvalida`, `TestHistorias_LasPrioridadesDelCodigoCoincidenConElCheckDeLaMigracion`, `TestOrdenar_PorPrioridadDespuesPorOrdenYDespuesPorNumero`, `TestMover_SubirIntercambiaConLaDeArribaYNumeraLasDeSuPrioridad`, `TestHistorias_ActualizarOrden`, `TestBacklog_OrdenadoPorPrioridadConSubirYBajar` | Cubierto |
| Product Backlog · Estado | HU-08 | `hu-08-estado-historia.feature` «Una historia recorre todos los estados hasta Hecho» · «Qué transiciones valen» · «Un cambio inválido deja la historia como estaba» | `TestEstados_EnElOrdenDelCamino`, `TestCambiarEstado_TransicionesValidas`, `TestCambiarEstado_ElCaminoCompletoHastaHecho` | Cubierto |
| Product Backlog · Story Points | HU-05, HU-16 | `hu-05-crear-item-backlog.feature` «Story points de la historia»; `hu-16-estimar-historia.feature` «Valores de la escala» | `TestNueva_StoryPointsDeLaEscala`, `TestNueva_StoryPointsInvalidos`, `TestEscalaStoryPoints_EsLaAcordada` | Cubierto |
| Product Backlog · Criterios de aceptación | HU-07 | `hu-07-criterios-aceptacion.feature` «Agregar el primer criterio de la historia» · «Cuándo una historia está completa» · «Una historia Hecha no admite cambios en sus criterios» | `TestProgreso_TieneCriterios`, `TestProgreso_Completo`, `TestAgregarCriterio_*`, `TestNuevoCriterio_*` | Cubierto |

## 3. Gestión de Sprints

| Requerimiento | Historia | Escenarios BDD | Tests | Estado |
|---|---|---|---|---|
| Sprints · Crear Sprints | HU-11 | `hu-11-crear-sprint.feature` «Crear el primer sprint del proyecto» · «No se superpone con el último sprint» · «Fechas en el borde del proyecto» | `TestNuevo_FechaDeFinPosteriorALaDeInicio`, `TestNuevo_InformaTodosLosCamposInvalidosJuntos`, `TestCrearSprint_DatosValidos_*` | Cubierto |
| Sprints · Definir un Sprint Goal | HU-11 | `hu-11-crear-sprint.feature` «El Sprint Goal es obligatorio» | `TestNuevo_LargoDelSprintGoal`, `TestNuevo_QuitaLosEspaciosAlrededorDelSprintGoal` | Cubierto |
| Sprints · Asignar historias a un Sprint | HU-12 | `hu-12-asignar-historias.feature` «Asignar una historia lista» · «Historias que no se pueden asignar» · «Quitar una historia del sprint» | `TestAsignarA_HistoriaLista`, `TestAsignarA_Errores`, `TestAsignar_HistoriaLista_QuedaGuardadaEnElSprint` | Cubierto |
| Sprints · Registrar historias completadas | HU-13 | `hu-13-completar-historias.feature` «Completar una historia del sprint activo» · «Faltan criterios» | `TestCompletar_HistoriaDelSprintActivo`, `TestCompletar_Errores`, `TestCompletar_SprintVencidoSinCerrar` | Cubierto |
| Sprints · Cerrar un Sprint | HU-14 | `hu-14-cerrar-sprint.feature` «Cerrar el sprint activo guarda su resultado» · «Las historias no completadas vuelven al backlog» · «La foto no cambia después del cierre» | `TestCerrar_SprintActivoEnCualquierDia`, `TestCerrar_Errores`, `TestCerrar_IgnoraLaHora` | Cubierto |
| Sprints · Consultar Sprints anteriores | HU-15 | `hu-15-sprints-anteriores.feature` «Ver los sprints anteriores con su resultado» · «Ver el detalle de un sprint cerrado» | `TestResumirSprints_OrdenadosConElResultadoDeLosCerrados`, `TestCumplimiento_CompletadosSobrePlanificados`, `TestListaDeSprints_MuestraElResultadoDeLosCerrados` | Cubierto |

## 4. Estimación y Planning Poker

| Requerimiento | Historia | Escenarios BDD | Tests | Estado |
|---|---|---|---|---|
| Estimación · Estimar historias mediante Story Points | HU-16, HU-16b | `hu-16-estimar-historia.feature` «Estimar una historia sin estimar» · «Re-estimar reemplaza el valor anterior»; `hu-16b-historial-estimaciones.feature` «Re-estimar conserva el valor anterior» | `TestEstimar_ValoresDeLaEscala`, `TestEstimar_Valido_GuardaLosStoryPoints`, `TestEstimar_ReemplazaLaEstimacionAnterior`, `TestNuevaReestimacion_ConUnValorDistinto_AnotaElCambio` | Cubierto |
| Estimación · Registrar las estimaciones individuales (Planning Poker) | HU-17, HU-18 | — | — | Pendiente: HU-17 y HU-18, Sprint 4 |
| Estimación · Mantenerlas ocultas hasta finalizar la votación | HU-18 | — | — | Pendiente: HU-18, Sprint 4 |
| Estimación · Mostrar las estimaciones realizadas | HU-19 | — | — | Pendiente: HU-19, Sprint 4 |
| Estimación · Detectar diferencias entre estimaciones | HU-20 | — | — | Pendiente: HU-20, Sprint 4 |
| Estimación · Realizar nuevas rondas | HU-21 | — | — | Pendiente: HU-21, Sprint 4 |
| Estimación · Registrar la estimación acordada | HU-22 | — | — | Pendiente: HU-22, Sprint 4 |

## 5. Registro de esfuerzo

| Requerimiento | Historia | Escenarios BDD | Tests | Estado |
|---|---|---|---|---|
| Esfuerzo · Integrante | HU-24, HU-25 | `hu-24-registro-esfuerzo.feature` «Un integrante dado de baja no registra horas»; `hu-25-descomponer-tareas.feature` «Un integrante dado de baja no puede ser responsable» | `TestNuevoRegistro_Validaciones`, `TestRegistrarEsfuerzo_DatosInvalidos` | Cubierto |
| Esfuerzo · Fecha | HU-24 | `hu-24-registro-esfuerzo.feature` «Fecha del registro» · «Total del día» | `TestNuevoRegistro_TotalDelDia`, `TestRegistrarEsfuerzo_TotalDelDiaEntreHistorias` | Cubierto |
| Esfuerzo · Actividad realizada | HU-24 | `hu-24-registro-esfuerzo.feature` «Registrar horas en la historia» | `TestNuevoRegistro_Validaciones`, `TestRegistros_LaBaseRechazaDatosInvalidos` | Cubierto |
| Esfuerzo · Horas trabajadas | HU-24, HU-26 | `hu-24-registro-esfuerzo.feature` «Horas de un registro» · «Las horas reales suman todos los registros de la historia»; `hu-26-registros-esfuerzo.feature` «Corregir un registro propio» | `TestNuevoRegistro_Validaciones`, `TestHorasReales_SumaLosRegistros`, `TestCorregirRegistro`, `TestPuedeCorregir` | Cubierto |
| Esfuerzo · Comparar el esfuerzo estimado con el real | HU-26, HU-33 | `hu-26-registros-esfuerzo.feature` «Ver horas estimadas y reales por historia»; `hu-33-desviacion.feature` «Una historia que llevó más horas de las estimadas» | `TestPaginaDeRegistros_FiltrosListaYHorasPorHistoria`, `TestCalcularDesviacion`, `TestPaginaDeHoras_MuestraLaDesviacion` | Cubierto |

## 6. Gestión de defectos

| Requerimiento | Historia | Escenarios BDD | Tests | Estado |
|---|---|---|---|---|
| Defectos · Descripción | HU-27 | `hu-27-registrar-defecto.feature` «Registrar un defecto de una historia» · «Datos inválidos de un defecto» | `TestNuevo_Validaciones`, `TestRegistrarDefecto_DatosInvalidos` | Cubierto |
| Defectos · Severidad | HU-27, HU-29 | `hu-27-registrar-defecto.feature` «Datos inválidos de un defecto»; `hu-29-consultar-defectos.feature` «Filtrar por severidad, estado y sprint» | `TestNuevo_Validaciones`, `TestConsultarDefectos`, `TestPaginaDeDefectos_FiltrosYCantidadPorSeveridad` | Cubierto |
| Defectos · Estado | HU-28 | `hu-28-estado-defecto.feature` «Llevar un defecto hasta Cerrado» · «Reabrir un defecto resuelto» · «Transiciones no permitidas» | `TestNuevo_ArrancaAbierto`, `TestCambiarEstadoDeDefecto`, `TestCambiarEstadoDeDefecto_CambioNoPermitido_Responde409` | Cubierto |
| Defectos · Historia relacionada | HU-27 | `hu-27-registrar-defecto.feature` «Registrar un defecto de una historia» · «Registrar un defecto sin historia» | `TestNuevo_SinHistoria`, `TestRegistrarDefecto_GuardaAbiertoYSeVeEnElProyecto` | Cubierto |
| Defectos · Sprint de detección | HU-27 | `hu-27-registrar-defecto.feature` «El sprint y la historia tienen que ser del proyecto» | `TestRegistrarDefecto_SprintDeOtroProyecto` | Cubierto |
| Defectos · Sprint de resolución | HU-28 | `hu-28-estado-defecto.feature` «Resolver en el mismo sprint en que se detectó» | `TestCambiarEstadoDeDefecto`, `TestDefectos_BuscarPorIDYActualizar` | Cubierto |

## 7. Métricas

| Requerimiento | Historia | Escenarios BDD | Tests | Estado |
|---|---|---|---|---|
| Métricas · Story Points planificados | HU-30 | `hu-30-sp.feature` «Planificados y completados de un sprint» · «El total del proyecto es la suma de sus sprints» | `TestCalcularStoryPoints_PlanificadosYCompletados`, `TestSumarStoryPoints_TotalDelProyecto` | Cubierto |
| Métricas · Story Points completados | HU-30 | `hu-30-sp.feature` «Planificados y completados de un sprint» · «Las historias sin estimar suman 0 y se avisa» | `TestCalcularStoryPoints_PlanificadosYCompletados`, `TestCalcularStoryPoints_HistoriasSinEstimarSumanCeroYSeAvisa` | Cubierto |
| Métricas · Velocidad del equipo | HU-31 | `hu-31-velocidad.feature` «Promedio de los últimos sprints cerrados» · «Proyecto sin sprints cerrados» · «Los sprints activos o planificados no cuentan» | `TestCalcularVelocidad_PromediaLosUltimosSprintsCerrados`, `TestCalcularVelocidad_IgnoraSprintsActivosYPlanificados`, `TestCalcularVelocidad_SinSprintsCerradosDevuelveCeroYAvisa` | Cubierto |
| Métricas · Horas estimadas | HU-23 | `hu-23-horas-estimadas.feature` «Cargar horas estimadas en una historia sin tareas» · «Historia con tareas» | `TestValidarHoras_MayoresACeroConDecimales`, `TestHorasEstimadasDeHistoria_SumaLasTareas` | Cubierto |
| Métricas · Horas reales | HU-24 | `hu-24-registro-esfuerzo.feature` «Las horas reales suman todos los registros de la historia» | `TestHorasReales_SumaLosRegistros` | Cubierto |
| Métricas · Desviación entre esfuerzo estimado y real | HU-33 | `hu-33-desviacion.feature` «Una historia que llevó menos horas de las estimadas» · «Desviación por sprint y por proyecto» | `TestCalcularDesviacion`, `TestSumarDesviaciones`, `TestPaginaDeHoras_SinEstimadas_NoAplica` | Cubierto |
| Métricas · Porcentaje de historias completadas | HU-32 | `hu-32-porcentaje.feature` «Porcentaje de un sprint» · «Porcentaje del proyecto con varios sprints» | `TestCalcularPorcentajeCompletadas_HechasSobrePlanificadas`, `TestSumarPorcentajes_TotalDelProyecto` | Cubierto |
| Métricas · Cantidad de defectos detectados | HU-34 | `hu-34-defectos.feature` «Detectados y resueltos por sprint» | `TestContarDefectos_PorSprintYTotal`, `TestResumenDeDefectos` | Cubierto |
| Métricas · Cantidad de defectos resueltos | HU-34 | `hu-34-defectos.feature` «Detectados y resueltos por sprint» · «Filtrar por severidad» | `TestContarDefectos_ResueltoSinSprintNoCuenta`, `TestContarDefectos_FiltraPorSeveridad` | Cubierto |

## 8. Dashboard

| Requerimiento | Historia | Escenarios BDD | Tests | Estado |
|---|---|---|---|---|
| Dashboard · Estado del proyecto | HU-35, HU-04 | `hu-35-panel.feature` «Panel con datos» · «Panel de un proyecto recién creado»; `hu-04-estado-proyecto.feature` «Proyecto en curso con su sprint activo» | `TestPanel_ConDatos`, `TestPanel_SinDatos`, `TestPanel_SinDatos_MuestraLasSeisTarjetas` | Cubierto |
| Dashboard · Representaciones gráficas de métricas | HU-36 | — | — | Pendiente: HU-36, Sprint 4 |

## 9. Reportes

| Requerimiento | Historia | Escenarios BDD | Tests | Estado |
|---|---|---|---|---|
| Reportes · Reporte de un Sprint | HU-37 | `hu-37-reporte-sprint.feature` «Las métricas del reporte de un sprint cerrado» · «Reporte de un sprint que todavía no se cerró» | `TestReporteDeSprint_Cerrado`, `TestReporteDeSprint_Abierto`, `TestReporteDeSprint_MuestraElReporte` | Cubierto |
| Reportes · Reporte de un proyecto | HU-38 | — | — | Pendiente: HU-38 |
| Reportes · Historias planificadas y completadas | HU-37 | `hu-37-reporte-sprint.feature` «El reporte conserva la foto del cierre» · «Las historias del reporte con sus horas» | `TestReporteDeSprint_ConservaLaFoto`, `TestReporteDeSprint_Cerrado` | Cubierto |
| Reportes · Estimaciones | HU-37 | `hu-37-reporte-sprint.feature` «Las historias del reporte con sus horas» | `TestReporteDeSprint_Cerrado`, `TestReporteDeSprint_MuestraElReporte` | Cubierto |
| Reportes · Esfuerzo registrado | HU-37 | `hu-37-reporte-sprint.feature` «Las historias del reporte con sus horas» | `TestReporteDeSprint_Cerrado`, `TestReporteDeSprint_MuestraElReporte` | Cubierto |
| Reportes · Métricas | HU-37 | `hu-37-reporte-sprint.feature` «Las métricas del reporte de un sprint cerrado» | `TestReporteDeSprint_Cerrado`, `TestReporteDeSprint_VelocidadHastaEseSprint` | Cubierto |
| Reportes · Defectos | HU-37 | `hu-37-reporte-sprint.feature` «Las métricas del reporte de un sprint cerrado» · «Reporte de un sprint sin historias ni defectos» | `TestReporteDeSprint_Defectos`, `TestReporteDeSprint_MuestraLosDefectos`, `TestReporteDeSprint_SinDefectos` | Cubierto |
| Reportes · Exportar el informe a PDF | HU-39 | — | — | Pendiente: HU-39, Sprint 4 |

## 10. Mínimo de pruebas unitarias que pide la consigna (TDD)

La consigna pide pruebas unitarias en Go para, como mínimo, estas cuatro cosas. Cada una se desarrolla con RED → GREEN en
commits separados (la evidencia es el historial del repositorio) y se mide con el umbral de cobertura de TEC-06.

| Requerimiento | Historia | Escenarios BDD | Tests | Estado |
|---|---|---|---|---|
| TDD · Cálculo de métricas | HU-30 a HU-34 | `hu-30-sp.feature`, `hu-31-velocidad.feature`, `hu-32-porcentaje.feature`, `hu-33-desviacion.feature`, `hu-34-defectos.feature` | `TestCalcularStoryPoints_PlanificadosYCompletados`, `TestCalcularVelocidad_PromediaLosUltimosSprintsCerrados`, `TestCalcularPorcentajeCompletadas_HechasSobrePlanificadas`, `TestCalcularDesviacion`, `TestContarDefectos_PorSprintYTotal` | Cubierto |
| TDD · Cálculos de estimación | HU-16, HU-16b, HU-23, HU-25 | `hu-16-estimar-historia.feature` «Valores de la escala»; `hu-23-horas-estimadas.feature` «Historia con tareas»; `hu-16b-historial-estimaciones.feature` «Re-estimar conserva el valor anterior» | `TestEstimar_ValoresDeLaEscala`, `TestEscalaStoryPoints_EsLaAcordada`, `TestHorasEstimadasDeHistoria_SumaLasTareas`, `TestParsearHoras_AceptaComaOPuntoDecimal`, `TestNuevaReestimacion_ConUnValorDistinto_AnotaElCambio` | Cubierto: falta la parte de Planning Poker (HU-20, Sprint 4) |
| TDD · Reglas de negocio | HU-03, HU-08, HU-12, HU-13, HU-14 | `hu-03-registrar-integrantes.feature` «No puede haber dos Agile Enabler activos»; `hu-08-estado-historia.feature` «Qué transiciones valen»; `hu-12-asignar-historias.feature` «Historias que no se pueden asignar»; `hu-14-cerrar-sprint.feature` «No se cierra un sprint que no es el activo» | `TestNuevo_UnSoloAgileEnablerActivo`, `TestCambiarEstado_TransicionesValidas`, `TestAsignarA_Errores`, `TestCompletar_Errores`, `TestCerrar_Errores`, `TestPuedeEliminarse_HistoriaHecha_NoSeElimina` | Cubierto |
| TDD · Validaciones | HU-01, HU-03, HU-05, HU-23, HU-24 | `hu-01-crear-proyecto.feature` «Se indica cada campo que falló»; `hu-03-registrar-integrantes.feature` «Formato del email»; `hu-05-crear-item-backlog.feature` «Story points de la historia»; `hu-23-horas-estimadas.feature` «Horas inválidas» | `TestNuevo_InformaTodosLosCamposInvalidosJuntos`, `TestNuevo_FormatoDelEmail`, `TestNueva_StoryPointsInvalidos`, `TestValidarHoras_MayoresACeroConDecimales`, `TestNuevoRegistro_Validaciones` | Cubierto |
