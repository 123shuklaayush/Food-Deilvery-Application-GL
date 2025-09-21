# Food-Deilvery-Application-GL

Monorepo

- frontend/: Vite + React app
- server/: Go (Gin) API

Local dev

1. Backend

```bash
cd server
cp .env.example .env  # or set env variables
go mod tidy && go run ./cmd/server
```

2. Frontend

```bash
cd frontend
npm i
npm run dev
```

Environment

- Backend
  - AUTH0_ISSUER_BASE_URL
  - AUTH0_AUDIENCE
  - MONGODB_CONNECTION_STRING
  - CLOUDINARY_CLOUD_NAME
  - CLOUDINARY_API_KEY
  - CLOUDINARY_API_SECRET
  - STRIPE_API_KEY
  - STRIPE_WEBHOOK_SECRET
  - FRONTEND_URL
- Frontend
  - VITE_AUTH0_CALLBACK_URL
  - VITE_API_BASE_URL
  - VITE_AUTH0_AUDIENCE
  - VITE_AUTH0_CLIENT_ID
  - VITE_AUTH0_DOMAIN

Deploy

- Backend (Render): import this repo (main) via render.yaml blueprint
- Frontend (Vercel): root directory `frontend/`
