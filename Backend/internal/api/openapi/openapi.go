// Package openapi serves the hand-maintained OpenAPI specification for the
// REST API. The spec is embedded so every deployment ships with its docs.
// When the swag CLI is available, `swag init` can regenerate docs from the
// annotations on the handlers; this embedded spec stays authoritative for the
// prototype.
package openapi

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var spec []byte

// RegisterRoutes exposes the specification and a small index page.
func RegisterRoutes(engine *gin.Engine) {
	engine.GET("/swagger", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(indexHTML))
	})
	engine.GET("/swagger/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", spec)
	})
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>PulseGraph API</title></head>
<body style="font-family:system-ui,sans-serif;max-width:720px;margin:40px auto;line-height:1.6">
  <h1>PulseGraph API</h1>
  <p>OpenAPI specification: <a href="/swagger/openapi.yaml">/swagger/openapi.yaml</a></p>
  <p>
    Health: <a href="/healthz">/healthz</a> &middot;
    Readiness: <a href="/readyz">/readyz</a> &middot;
    Metrics: <a href="/metrics">/metrics</a>
  </p>
  <p>Drop the specification into Swagger UI or Postman to explore the API.</p>
</body>
</html>
`
