# Bitely API

REST API for the Bitely iOS app. Built with **Go + PostgreSQL**, includes **JWT authentication**, migrations, and user-scoped authorization for shared recipes.

**iOS app repo:** [bitely-ios](https://github.com/thomaslgrega/bitely-ios)

---

## Features

- CRUD for recipes + ingredients
- PostgreSQL schema + migrations
- JWT auth (Bearer tokens)
- Protected endpoints enforce recipe ownership (`WHERE id = $1 AND user_id = $2`)
- Session restore support (`GET /me`)

Sign in with Email  
Sign in with Apple: **in progress**

---

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

---

### Sample `.env` file:
PORT=8080  
DATABASE_URL=postgres://postgres:password@localhost:5432/bitelyapi?sslmode=disable  
JWT_SECRET=your_long_random_secret
