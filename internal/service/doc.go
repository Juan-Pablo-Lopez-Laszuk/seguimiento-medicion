// Package service implementa los casos de uso: valida con el dominio y coordina los repositorios.
//
// Depende de interfaces de repositorio (no de Supabase directamente), así los tests usan
// los repositorios en memoria de internal/store/memory.
package service
