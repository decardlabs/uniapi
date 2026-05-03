# One API Frontend

> Modern is the only maintained and bundled frontend template.

> [!NOTE]
> Legacy themes (`default`, `air`, `berry`) are automatically mapped to `modern`.

## Development

1. Start backend from project root:

```bash
go run main.go
```

2. Start Modern frontend from project root:

```bash
make dev-modern
```

## URLs

- Modern frontend: `http://localhost:3001`
- Backend API: `http://localhost:3000`

The Modern dev server proxies `/api` requests to the backend.

## Build

Build Modern frontend from project root:

```bash
make build-frontend-modern
```
