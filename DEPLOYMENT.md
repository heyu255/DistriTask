# Distritask Deployment Guide

This app consists of:
1. **Next.js Frontend** - Deploy to Vercel
2. **Go Backend Services** - Deploy to Railway/Render (Manager, Worker, Dashboard Server)
3. **Redis** - Use Upstash or Redis Cloud

## Architecture

- **Manager Service** (Port 8080) - Handles task submission
- **Worker Services** (Multiple instances) - Process tasks from Redis stream
- **Dashboard Server** (Port 8081) - WebSocket server for real-time updates
- **Redis** - Task queue and pub/sub for updates

## Deployment Steps

### 1. Deploy Redis

**Option A: Upstash (Recommended)**
1. Go to https://upstash.com
2. Create a new Redis database
3. Copy the Redis URL (format: `redis://default:password@host:port`)

**Option B: Redis Cloud**
1. Go to https://redis.com/cloud
2. Create a free database
3. Copy the connection string

### 2. Deploy Go Services

**Using Railway (Recommended)**

1. Install Railway CLI: `npm i -g @railway/cli`
2. Login: `railway login`
3. Create new project: `railway init`
4. For each service (manager, worker, dashboard):
   ```bash
   cd cmd/manager  # or worker, dashboard
   railway init
   railway up
   ```
5. Set environment variables in Railway dashboard:
   - `REDIS_URL` - Your Redis connection string
   - `PORT` - Railway will auto-assign, but you can set it

**Using Render**

1. Go to https://render.com
2. Create new Web Service for each Go service
3. Build command: `go build -o distritask ./cmd/manager` (adjust for each service)
4. Start command: `./distritask`
5. Set environment variables:
   - `REDIS_URL`
   - `PORT`

### 3. Update Go Services for Production

The Go services currently use `localhost:6379` for Redis. Update them to use environment variables:

**Update `cmd/manager/main.go`, `cmd/worker/main.go`, `cmd/dashboard/main.go`:**

```go
redisAddr := os.Getenv("REDIS_URL")
if redisAddr == "" {
    redisAddr = "localhost:6379" // fallback for local dev
}
rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
```

### 4. Deploy Next.js Frontend to Vercel

1. Push your code to GitHub
2. Go to https://vercel.com
3. Import your repository
4. Set environment variables:
   - `NEXT_PUBLIC_API_URL` - Your Manager service URL (e.g., `https://your-manager.railway.app`)
   - `NEXT_PUBLIC_WS_URL` - Your Dashboard server WebSocket URL (e.g., `wss://your-dashboard.railway.app/ws`)
5. Deploy!

### 5. Environment Variables Summary

**Frontend (Vercel):**
- `NEXT_PUBLIC_API_URL` - Manager service HTTP URL
- `NEXT_PUBLIC_WS_URL` - Dashboard server WebSocket URL

**Backend Services (Railway/Render):**
- `REDIS_URL` - Redis connection string
- `PORT` - Service port (usually auto-assigned)

## Quick Start (Local Development)

1. Start Redis: `redis-server` (or use Docker: `docker run -p 6379:6379 redis`)
2. Start Manager: `go run cmd/manager/main.go`
3. Start Workers: `go run cmd/worker/main.go` (run multiple instances)
4. Start Dashboard Server: `go run cmd/dashboard/main.go`
5. Start Frontend: `npm run dev`

## Production Checklist

- [ ] Redis deployed and accessible
- [ ] Manager service deployed with `REDIS_URL` set
- [ ] Worker services deployed (at least 1 instance)
- [ ] Dashboard server deployed with WebSocket support
- [ ] Frontend deployed with `NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_WS_URL` set
- [ ] All services can connect to Redis
- [ ] WebSocket connections work (check CORS if needed)

