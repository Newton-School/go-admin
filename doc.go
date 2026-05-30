// Package admin provides a Django-inspired admin SDK for Go applications.
//
// The package exposes a standard net/http handler. Host applications register
// apps and typed resources, then mount the handler behind their own auth
// middleware.
package admin
