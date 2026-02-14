# cmd/

Application entry points for the platform-go project.

## Table of Contents

1. [Overview](#overview)
2. [Structure](#structure)
3. [Subdirectories](#subdirectories)

---

## Overview

Contains main applications for the platform. Each subdirectory has its own `main.go` that serves as an executable entry point.

## Structure

```
cmd/
└─ api/         - HTTP API server entry point
```

## Subdirectories

### api/

HTTP REST API server implementation.

- Starts the HTTP server with REST API endpoints
- Listens on port 8080 by default
- Handles all RESTful API requests
- Integrated with PostgreSQL database