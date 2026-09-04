# Bitely API

REST API for the Bitely iOS app. Built with **Go + PostgreSQL**, includes **JWT authentication**, SQL migrations (**golang-migrate**), and user-scoped authorization for shared recipes.

**iOS app repo:** [bitely-ios](https://github.com/thomaslgrega/bitely-ios)

## Hosted
- **API (Render, Docker):** https://bitelyapi-docker.onrender.com
- **Database:** Neon (Postgres)

Health check: `GET /health`

## Features
- CRUD for recipes + ingredients
- Recipe images uploaded direct to Cloudflare R2 and served from the bucket
- PostgreSQL schema + SQL migrations (golang-migrate)
- JWT auth (Bearer tokens)
- Protected endpoints enforce recipe ownership (e.g., `WHERE id = $1 AND user_id = $2`)
- Session restore support (`GET /me`)

## Auth options
- Email sign-in  
- Sign in with Apple: **in progress**

## Routes

### Public
- `GET /recipes` — the feed: recipe summaries for a client with nothing to narrow by, most recently shared first. Up to 50; `?limit=` lowers that (and lowers a name query's 50 too, but is refused alongside `category`, which is uncapped). See ADR-0005.
- `GET /recipes?category=Dessert` — list recipe summaries by category
- `GET /recipes?name=shakshuka` — list recipe summaries whose **name** matches the query, closest first. The match is fuzzy, so a misspelling still finds the recipe; it never looks at ingredients (that is `POST /recipes/match`). Composable with `category` to search within one. Up to 50 results. See ADR-0004.
- `GET /recipes/{id}` — get full recipe details (includes ingredients). A recipe with an image carries `image_url`, composed from the stored key and `R2_PUBLIC_BASE_URL`; a recipe without one carries no `image_url` at all. Writes name the image by `image_key` instead, on `POST /recipes` and on the image sub-resource. See ADR-0006.
- `POST /recipes/match` — match a pantry against the recipe corpus. Body is a JSON list of raw ingredient strings, e.g. `["chicken", "rice", "onion"]`; the response is up to 50 recipe cards ordered by coverage, each with its matched ingredients, missing ingredients, and coverage. See `docs/ingredient-matching-algorithm.md`.

### Auth
- `POST /auth/register`
- `POST /auth/login`
- `GET /me` — current user (requires auth)

### Protected (requires `Authorization: Bearer <token>`)
- `POST /recipes` — create shared recipe. A recipe with an image names it by `image_key`, the key returned by `POST /recipes/images`; the server checks the staged object's size and type, then promotes it to a key it derives itself. See ADR-0006.
- `POST /recipes/images` — mint a presigned upload for one recipe image. Body is `{"content_type":"image/jpeg","content_length":123456}`; the response is `{"upload_url","key","expires_at"}`. `PUT` the bytes straight to `upload_url` with exactly that `Content-Type` and `Content-Length`, then send the `key` back as `image_key` when sharing. JPEG and PNG only, 5MB, and the URL lives five minutes. Rate limited per user.
- `GET /me/recipes` — list my shared recipes
- `PUT /recipes/{id}` — update my recipe. It replaces every field it carries, so an omitted field is a deleted one. It carries no image: an `image_key` here is refused with a `400` naming the endpoint below. See ADR-0006.
- `PUT /recipes/{id}/image` — set my recipe's image. Body is `{"image_key":"incoming/<uuid>"}`; the response is `{"image_url":"..."}` for the key the server promoted it to. A save that changes both text and image calls this first and abandons the save if it fails.
- `DELETE /recipes/{id}/image` — remove my recipe's image. Answers `204` whether or not there was one, so a retried save cannot fail on its second attempt.
- `DELETE /recipes/{id}` — delete my recipe

---

## Running Locally (Go)

### 1. Configure environment variables
Create a `.env` file:

```env
PORT=8080
DATABASE_URL=postgres://postgres:password@localhost:5432/bitelyapi?sslmode=disable
JWT_SECRET=your_secret
R2_ACCOUNT_ID=your_cloudflare_account_id
R2_ACCESS_KEY_ID=your_r2_access_key_id
R2_SECRET_ACCESS_KEY=your_r2_secret_access_key
R2_BUCKET=bitely-images
R2_PUBLIC_BASE_URL=https://pub-<hash>.r2.dev
```

`scripts/provision-r2.sh` walks through the Cloudflare and Render side of that:
it creates the bucket, enables the Public Development URL, mints an API token
scoped to the one bucket, sets the lifecycle rule that expires abandoned
uploads, and writes the five values into `.env`.

### 2. Start Postgres
Start a local Postgres instance (local install, Docker, etc.).

### 3. Run migrations (golang-migrate)
```bash
migrate -path migrations -database "$DATABASE_URL" up
```

### 4. Start the API
```bash
go run ./cmd/server
```

API will be available at `http://localhost:8080`.

---

## Running Locally (Docker)

### Recommended (Makefile)

Build + Run:
```bash
DATABASE_URL="YOUR_DATABASE_URL" JWT_SECRET="YOUR_SECRET" \
  R2_ACCOUNT_ID="YOUR_ACCOUNT_ID" R2_ACCESS_KEY_ID="YOUR_KEY_ID" \
  R2_SECRET_ACCESS_KEY="YOUR_SECRET_KEY" R2_BUCKET="bitely-images" \
  R2_PUBLIC_BASE_URL="https://pub-<hash>.r2.dev" make docker-up
```

Verify:
```bash
curl http://localhost:8080/health
```

Stop (if needed):
```bash
make docker-down
```

### Alternative (raw Docker commands)

Build:
```bash
docker build -t bitely-api .
```

Run:
```bash
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="YOUR_DATABASE_URL" \
  -e JWT_SECRET="YOUR_SECRET" \
  -e R2_ACCOUNT_ID="YOUR_ACCOUNT_ID" \
  -e R2_ACCESS_KEY_ID="YOUR_KEY_ID" \
  -e R2_SECRET_ACCESS_KEY="YOUR_SECRET_KEY" \
  -e R2_BUCKET="bitely-images" \
  -e R2_PUBLIC_BASE_URL="https://pub-<hash>.r2.dev" \
  bitely-api
```

Verify:
```bash
curl http://localhost:8080/health
```

## Deployment Notes (Render + Neon)
- API is containerized with Docker and deployed to Render.
- Render configuration is managed via environment variables (`DATABASE_URL`, `JWT_SECRET`, the `R2_*` values, etc.).
- Neon hosts the production PostgreSQL database.
- Base URL: https://bitelyapi-docker.onrender.com

## Roadmap
- Sign in with Apple
- Pagination + improved search
- Basic observability (structured logs / request IDs)