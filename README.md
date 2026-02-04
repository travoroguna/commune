# Commune

A full-stack application starter template with Go backend and React frontend.

## Tech Stack

### Backend
- **Go** - Modern, fast backend language
- **GORM** - ORM library for Go
- **gormmigrate** - Database migrations
- **olivere/vite** - Vite integration for serving frontend

### Frontend
- **React** - UI library
- **TypeScript** - Type-safe JavaScript
- **Vite** - Fast build tool and dev server
- **Tailwind CSS** - Utility-first CSS framework
- **shadcn/ui** - Re-usable component library

## Project Structure

```
commune/
├── backend/           # Go backend application
│   ├── main.go       # Main application file
│   ├── go.mod        # Go dependencies
│   └── go.sum
├── frontend/          # React frontend application
│   ├── src/          # Source files
│   ├── public/       # Static assets
│   ├── package.json  # Node dependencies
│   └── vite.config.ts
└── scripts/          # Build and development scripts
    ├── dev.sh        # Development mode script
    └── build.sh      # Production build script
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- npm or yarn
- [Air](https://github.com/air-verse/air) (for auto-reloading Go backend in dev mode)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/travoroguna/commune.git
cd commune
```

2. Install backend dependencies:
```bash
cd backend
go mod download
cd ..
```

3. Install frontend dependencies:
```bash
cd frontend
npm install
cd ..
```

## Development Mode

In development mode, the frontend and backend run independently with hot-reloading:
- Vite dev server runs on port 5173 with Hot Module Replacement (HMR)
- Go backend runs on port 3000 with [Air](https://github.com/air-verse/air) for auto-rebuilding
- The Go backend proxies requests to Vite dev server for HMR

Air automatically watches for changes in your Go files and rebuilds/restarts the backend server when changes are detected.

To start both servers:

```bash
./scripts/dev.sh
```

The script will automatically install Air if it's not already installed.

Access the application at: http://localhost:3000

## Production Mode

In production mode, the frontend is built as static files and served by the Go backend.

### Build for production:

```bash
./scripts/build.sh
```

### Run the production server:

```bash
cd backend
MODE=production PORT=3000 ./commune
```

## API Endpoints

- `GET /api/health` - Health check endpoint
- `GET /api/users` - Get all users

## Environment Variables

- `MODE` - Set to `development` or `production` (default: `development`)
- `PORT` - Server port (default: `3000`)

## Development

### Adding a new migration

Edit `backend/main.go` and add a new migration to the `runMigrations` function:

```go
{
    ID: "202402041301",
    Migrate: func(tx *gorm.DB) error {
        // Your migration code
        return nil
    },
    Rollback: func(tx *gorm.DB) error {
        // Your rollback code
        return nil
    },
}
```

### Adding shadcn/ui components

The project includes a basic shadcn/ui setup with the Button component. You can add more components as needed by copying them from the [shadcn/ui documentation](https://ui.shadcn.com/).

## License

MIT