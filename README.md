# Distritask - Distributed Task Monitor

A real-time distributed task monitoring system with Next.js frontend and Go backend services.

## Architecture

- **Frontend**: Next.js dashboard with WebSocket real-time updates
- **Manager Service**: Handles task submission (Port 8080)
- **Worker Services**: Process tasks from Redis stream (multiple instances)
- **Dashboard Server**: WebSocket server for real-time updates (Port 8081)
- **Redis**: Task queue and pub/sub for status updates

## Quick Deploy to Vercel (Frontend Only)

For a quick deployment of just the frontend:

1. Push to GitHub
2. Go to [Vercel](https://vercel.com) and import your repo
3. Set environment variables:
   - `NEXT_PUBLIC_API_URL` - Your manager service URL
   - `NEXT_PUBLIC_WS_URL` - Your dashboard WebSocket URL (wss://...)
4. Deploy!

## Full Deployment

See [DEPLOYMENT.md](./DEPLOYMENT.md) for complete deployment instructions including backend services.

## Local Development

1. Start Redis: `docker run -p 6379:6379 redis` or `redis-server`
2. Start Manager: `go run cmd/manager/main.go`
3. Start Worker(s): `go run cmd/worker/main.go` (run multiple terminals)
4. Start Dashboard Server: `go run cmd/dashboard/main.go`
5. Start Frontend: `npm run dev`

## Environment Variables

**Frontend:**
- `NEXT_PUBLIC_API_URL` - Manager service HTTP URL
- `NEXT_PUBLIC_WS_URL` - Dashboard WebSocket URL

**Backend Services:**
- `REDIS_URL` - Redis connection string
- `PORT` - Service port (auto-assigned on most platforms)
- `ALLOWED_ORIGIN` - CORS allowed origin (for manager service)

