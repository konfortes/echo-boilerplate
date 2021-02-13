# echo-boilerplate

A Go [Echo](https://github.com/labstack/echo) boilerplate

## Development

### dotenv

echo-boilerplate uses dotenv for development environment. set environment variables in `.env` file

## Prometheus

echo-boilerplate has a built-in Prometheus middleware. Metrics can be scrape at `/metrics`

## Jaeger

Tracing is disabled by default. To enable it, set `TRACING_ENABLED` to true. To configure agent - set `JAEGER_AGENT_HOST` and `JAEGER_AGENT_PORT` environment variables