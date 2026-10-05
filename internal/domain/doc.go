// Package domain agrupa las entidades y reglas de negocio de la aplicación, un subpaquete por épica
// (project, backlog, sprint, estimation, effort, defect).
//
// El dominio no conoce ni la base de datos ni HTTP: sus funciones reciben datos simples y devuelven
// resultados o errores, por eso se desarrolla con TDD sin dependencias externas.
package domain
