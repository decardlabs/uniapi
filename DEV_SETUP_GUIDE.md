# one-api Project Environment Setup Guide

## 1. Install Dependencies
- Go (>=1.25)
- Node.js (>=16)
- yarn
- MySQL (if used)
- Redis (if used)
- git

## 2. Clone the Repository (already done)

## 3. Backend Setup
- Install Go dependencies:
  ```sh
  cd ~/Documents/uniapi
  go mod tidy
  ```
- Build backend:
  ```sh
  make build
  ```

## 4. Frontend Setup
- Install frontend dependencies:
  ```sh
  cd web/modern
  yarn install
  ```
- Build frontend:
  ```sh
  yarn build
  ```

## 5. Database Setup
- Ensure MySQL and Redis are running and accessible.
- Create and configure databases as needed.

## 6. Environment Variables
- Copy and edit example env files if present (e.g., `.env.example` → `.env`).
- Set DB, Redis, and other secrets as required.

## 7. Debugging Tools
- Install Delve for Go debugging:
  ```sh
  go install github.com/go-delve/delve/cmd/dlv@latest
  ```

## 8. Running Locally
- Start backend:
  ```sh
  ./one-api
  ```
- Start frontend (if needed for dev):
  ```sh
  yarn dev
  ```

## 9. Deployment
- Use systemd, Docker, or another process manager for production.
- Example systemd service file: `one-api.service`
- Reload systemd and enable service:
  ```sh
  sudo systemctl daemon-reload
  sudo systemctl enable one-api
  sudo systemctl start one-api
  ```

## 10. Testing
- Run tests:
  ```sh
  go test -race ./...
  make build-frontend-modern
  ```

---

# SSH Access
- Use your SSH key to connect: `ssh user@43.167.187.228`
- Ensure correct permissions for deployment and runtime directories.

# Notes
- Adjust paths and usernames as needed.
- For more details, see README.md and DESIGN.md.
