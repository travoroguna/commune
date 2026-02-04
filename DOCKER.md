# Docker Setup Guide

This guide explains how to run the Commune application using Docker Compose.

## Services Included

The Docker Compose setup includes the following services:

1. **Main Application** (Backend + Frontend)
   - Go backend serving the application on port 3000
   - React frontend built and served by the backend

2. **PostgreSQL** (Database)
   - Port: 5432
   - Database: commune
   - User: commune
   - Password: commune_password (change in production!)

3. **Valkey** (Redis-compatible cache)
   - Port: 6379
   - Used for caching and session management

4. **SeaweedFS** (Object Storage)
   - Master Server: Port 9333
   - Volume Server: Port 8080
   - Filer (File Management): Port 8888

## Quick Start

### Prerequisites

- Docker Engine 20.10 or higher
- Docker Compose V2 or higher

### Running the Application

1. Clone the repository:
```bash
git clone https://github.com/travoroguna/commune.git
cd commune
```

2. Start all services with Docker Compose:
```bash
docker compose up -d
```

3. View logs (optional):
```bash
docker compose logs -f app
```

4. Access the application:
   - **Main Application**: http://localhost:3000
   - **PostgreSQL**: localhost:5432
   - **Redis/Valkey**: localhost:6379
   - **SeaweedFS Master**: http://localhost:9333
   - **SeaweedFS Volume**: http://localhost:8080
   - **SeaweedFS Filer**: http://localhost:8888

### Stopping the Application

```bash
docker compose down
```

To also remove volumes (WARNING: this will delete all data):
```bash
docker compose down -v
```

## Configuration

### Environment Variables

You can customize the application by modifying environment variables in `docker-compose.yml`:

**Database Configuration:**
- `DB_HOST`: PostgreSQL host (default: postgres)
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_NAME`: Database name (default: commune)
- `DB_USER`: Database user (default: commune)
- `DB_PASSWORD`: Database password (default: commune_password)

**Redis Configuration:**
- `REDIS_HOST`: Redis host (default: redis)
- `REDIS_PORT`: Redis port (default: 6379)

**SeaweedFS Configuration:**
- `SEAWEEDFS_MASTER`: Master server address (default: seaweedfs-master:9333)
- `SEAWEEDFS_FILER`: Filer server address (default: seaweedfs-filer:8888)

### Production Deployment

For production use, make sure to:

1. Change all default passwords
2. Configure proper SSL/TLS certificates
3. Set up proper backup strategies for volumes
4. Use proper secrets management instead of environment variables
5. Configure firewall rules appropriately

### Volumes

The setup uses Docker volumes for data persistence:

- `postgres_data`: PostgreSQL database files
- `redis_data`: Redis/Valkey data
- `seaweedfs_master_data`: SeaweedFS master data
- `seaweedfs_volume_data`: SeaweedFS volume data (actual file storage)

## Troubleshooting

### Check Service Health

```bash
docker compose ps
```

### View Logs for a Specific Service

```bash
docker compose logs -f [service-name]
# Example: docker compose logs -f app
# Example: docker compose logs -f postgres
```

### Restart a Service

```bash
docker compose restart [service-name]
```

### Rebuild the Application

If you made changes to the code:

```bash
docker compose up --build -d app
```

### Database Migrations

Database migrations run automatically when the application starts. If you need to manually run migrations, you can:

```bash
docker compose exec app ./commune
```

## Using SeaweedFS

SeaweedFS provides a distributed object storage system. Here are some basic operations:

### Upload a File
```bash
curl -F file=@example.txt "http://localhost:9333/submit"
```

### Access via Filer (Recommended)
The Filer provides a POSIX-like file system interface:

```bash
# Upload
curl -X POST "http://localhost:8888/path/to/file.txt" -d "file content"

# Download
curl "http://localhost:8888/path/to/file.txt"

# List files
curl "http://localhost:8888/path/"
```

For more information, see the [SeaweedFS documentation](https://github.com/seaweedfs/seaweedfs/wiki).

## Development vs Production

- The Docker setup runs the application in **production mode** by default
- For development with hot-reloading, use the native setup described in the main README.md
- The Docker setup is optimized for production deployment and testing

## Network

All services run on a dedicated Docker network (`commune-network`) and can communicate with each other using service names as hostnames.

## Support

For issues or questions:
- Check the main [README.md](README.md) for general application documentation
- Review Docker Compose logs for error messages
- Ensure all services are healthy: `docker compose ps`
