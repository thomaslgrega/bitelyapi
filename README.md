# Bitely API

REST API for the Bitely iOS app. Built with **Go + PostgreSQL**, includes **JWT authentication**, SQL migrations (golang-migrate), and user-scoped authorization for shared recipes.

**iOS app repo:** [bitely-ios](https://github.com/thomaslgrega/bitely-ios)

### **Hosted:**
- **API:** Render - `https://bitelyapi.onrender.com`
- **Database:** Neon (PostgreSQL)

## Features

- CRUD for recipes + ingredients
- PostgreSQL schema + SQL migrations (golang-migrate)
- JWT auth (Bearer tokens)
- Protected endpoints enforce recipe ownership (e.g., `WHERE id = $1 AND user_id = $2`)
- Session restore support (`GET /me`)

### **Auth options:**
- Email sign-in  
- Sign in with Apple: **in progress**


## Routes

### Public
- `GET /recipes?category=Dessert` — list recipe summaries by category
- `GET /recipes/{id}` — get full recipe details (includes ingredients)

### Auth
- `POST /auth/register`
- `POST /auth/login`
- `GET /me` — current user (requires auth)

### Protected (requires `Authorization: Bearer <token>`)
- `POST /recipes` — create shared recipe
- `GET /me/recipes` — list my shared recipes
- `PUT /recipes/{id}` — update my recipe
- `DELETE /recipes/{id}` — delete my recipe

## Running Locally

### 1. Configure environment variables

Create a `.env` file:
```
PORT=8080
DATABASE_URL=postgres://postgres:password@localhost:5432/bitelyapi?sslmode=disable  
JWT_SECRET=your_long_random_secret
```

### 2. Start Postgres

Start a local Postgres instance (local install, Docker, etc.).

### 3. Run migrations (golang-migrate)
```bash
migrate -path migrations -database "$DATABASE_URL" up
```

### 4. Start the API
```bash
go run cmd/server/main.go
```

API will be available at `http://localhost:8080`.

## Deployment Notes (Render + Neon)
- Render hosts the Go API; configuration is managed via environment variables.
- Neon hosts the production PostgreSQL database.
- Hosted base URL: https://bitelyapi.onrender.com

## Roadmap
- Sign in with Apple
- Image upload for shared recipes (S3 / R2)
- Pagination + improved search
- Basic observability (structured logs / request IDs)