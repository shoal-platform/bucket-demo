# Text Vault (bucket-demo)

A small demo app for uploading, listing, and downloading text files stored in a Google Cloud Storage bucket.

- **backend/**: Go HTTP API that talks to Cloud Storage
- **frontend/**: Next.js 14 (App Router) UI

## Prerequisites

- Go 1.22 or newer
- Node.js 18 or newer and npm
- A Google Cloud project with a Cloud Storage bucket
- Credentials that can read and write objects in that bucket (for example the `Storage Object Admin` role on the bucket)

## Configuration

The backend reads its settings from environment variables. Copy the example file and fill in your values:

```bash
cp backend/.env.example backend/.env
```

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `BUCKET_NAME` | Yes | | Name of the Cloud Storage bucket |
| `PORT` | No | `8080` | Port the API listens on |
| `GOOGLE_APPLICATION_CREDENTIALS` | No | | Path to a service account key JSON file |

### Authenticating with Google Cloud

The backend uses [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials). Pick one:

- **Local user account:** run `gcloud auth application-default login` and leave `GOOGLE_APPLICATION_CREDENTIALS` unset.
- **Service account key:** set `GOOGLE_APPLICATION_CREDENTIALS` to the path of the key file.
- **Running on Google Cloud** (Cloud Run, GCE, GKE): the attached service account is used automatically.

Never commit `.env` files or key files. Both are listed in `.gitignore`.

## Running locally

### Backend

Go does not load `.env` files on its own, so export the variables into your shell first.

macOS / Linux / Git Bash:

```bash
cd backend
set -a; source .env; set +a
go mod tidy
go run .
```

Windows PowerShell:

```powershell
cd backend
Get-Content .env | Where-Object { $_ -match '^\s*[^#].*=' } | ForEach-Object {
  $name, $value = $_ -split '=', 2
  Set-Item "env:$($name.Trim())" $value.Trim()
}
go mod tidy
go run .
```

The API listens on the port set in `PORT`, or `8080` if it is not set. The examples below assume the default. Check it with:

```bash
curl http://localhost:8080/health
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:3000`.

The frontend uses the `PORT` environment variable and falls back to `3000`. Set it in your shell when starting the server. Next.js picks the port before it loads `.env` files, so putting `PORT` in a `.env` file has no effect.

```bash
PORT=4000 npm run dev          # macOS / Linux / Git Bash
$env:PORT=4000; npm run dev    # Windows PowerShell
```

The frontend calls the backend through relative `/api/*` paths (`/api/items`, `/api/upload`, `/api/download`). These need to be forwarded to the backend, for example with a rewrite in `frontend/next.config.js`. Reading the backend URL from an environment variable keeps it in step with the backend's `PORT`:

```js
const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'

module.exports = {
  async rewrites() {
    return [{ source: '/api/:path*', destination: `${backendUrl}/:path*` }]
  },
}
```

## API reference

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health` | Returns `{"status":"ok"}` |
| `GET` | `/items` | Lists every object in the bucket as `[{ name, size, updated, contentType }]` |
| `POST` | `/upload` | Multipart form upload, field name `file`, max 10 MB. Returns `{ "name": "..." }` |
| `GET` | `/download?name=<file>` | Downloads the named object as an attachment |

Examples:

```bash
curl http://localhost:8080/items
curl -F "file=@notes.txt" http://localhost:8080/upload
curl -OJ "http://localhost:8080/download?name=notes.txt"
```

Notes:

- Uploaded files are stored under their original base name. Uploading a file with an existing name overwrites it.
- All objects are stored and served as `text/plain; charset=utf-8`.
- CORS is open to all origins (`*`). Restrict this before deploying anywhere public.

## Project structure

```
.
├── backend/
│   ├── go.mod
│   ├── main.go          # HTTP server and Cloud Storage handlers
│   └── .env.example
└── frontend/
    ├── package.json
    └── app/
        ├── layout.js
        ├── page.js      # Upload form and file list
        └── globals.css
```

## Scripts

| Location | Command | Description |
| --- | --- | --- |
| `backend/` | `go run .` | Start the API |
| `backend/` | `go build -o textvault .` | Build a binary |
| `frontend/` | `npm run dev` | Start the dev server on `$PORT` (default 3000) |
| `frontend/` | `npm run build` | Production build |
| `frontend/` | `npm run start` | Serve the production build on `$PORT` (default 3000) |