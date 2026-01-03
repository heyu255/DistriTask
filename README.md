# Distritask - Distributed Task Monitoring System

A real-time distributed task monitoring system with a modern Next.js frontend and Go backend services. Tasks are distributed across multiple workers using Redis streams, with real-time status updates via WebSocket.

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Next.js Frontend                         │
│              (Vercel - https://distri-task.vercel.app)      │
│                                                              │
│  - Real-time dashboard with WebSocket connection          │
│  - Task submission via HTTP POST                           │
│  - Live status updates from workers                        │
└─────────────────────────────────────────────────────────────┘
                          ↕ HTTP/WebSocket
┌─────────────────────────────────────────────────────────────┐
│                    Go Backend Services                      │
│                  (Railway - Railway.app)                    │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   Manager    │  │   Worker     │  │  Dashboard   │    │
│  │  (Port 8080) │  │  (Multiple)  │  │  (Port 8081) │    │
│  │              │  │              │  │              │    │
│  │ - Accepts    │  │ - Processes  │  │ - WebSocket  │    │
│  │   task POST   │  │   tasks     │  │   server     │    │
│  │ - Enqueues    │  │ - Updates    │  │ - Broadcasts │    │
│  │   to Redis    │  │   status    │  │   updates    │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
└─────────────────────────────────────────────────────────────┘
                          ↕ Redis Streams & Pub/Sub
┌─────────────────────────────────────────────────────────────┐
│                    Redis (Upstash)                          │
│                                                              │
│  - Task Queue (Redis Streams)                               │
│  - Status Updates (Pub/Sub)                                │
└─────────────────────────────────────────────────────────────┘
```

### Components

- **Frontend (Next.js)**: Real-time dashboard displaying task status across workers
- **Manager Service**: HTTP API that accepts task submissions and enqueues them to Redis
- **Worker Service**: Processes tasks from Redis streams (can run multiple instances)
- **Dashboard Server**: WebSocket server that broadcasts task status updates to connected clients
- **Redis**: Message broker for task queue (streams) and real-time updates (pub/sub)

## 🚀 Quick Start

### Prerequisites

- Go 1.25.5+
- Node.js 20+
- Redis (local or Upstash cloud)
- Railway account (for backend deployment)
- Vercel account (for frontend deployment)

### Local Development

1. **Start Redis**:
   ```bash
   # Using Docker
   docker run -p 6379:6379 redis
   
   # Or using local Redis
   redis-server
   ```

2. **Start Manager Service**:
   ```bash
   cd distritask
   go run cmd/manager/main.go
   # Runs on http://localhost:8080
   ```

3. **Start Worker Service** (run in separate terminals):
   ```bash
   go run cmd/worker/main.go
   # Can run multiple instances for parallel processing
   ```

4. **Start Dashboard WebSocket Server**:
   ```bash
   go run cmd/dashboard/main.go
   # Runs on ws://localhost:8081
   ```

5. **Start Frontend**:
   ```bash
   npm install
   npm run dev
   # Runs on http://localhost:3000
   ```

6. **Open the Dashboard**:
   - Navigate to `http://localhost:3000/dashboard`
   - Click "Deploy Node" to submit a task
   - Watch tasks appear in real-time across worker columns

## 📦 Deployment

### Backend Services (Railway)

1. **Create Railway Project**:
   - Go to [Railway](https://railway.app)
   - Create a new project
   - Connect your GitHub repository

2. **Add Services**:
   - Create 3 services: `manager`, `worker`, and `dashboard`
   - For each service, set the "Root Directory" to `.` (project root)

3. **Configure Build Settings**:
   - **Manager**: Use `railway.manager.toml` (Config as Code)
   - **Worker**: Use `railway.worker.toml` (Config as Code)
   - **Dashboard**: Use `railway.dashboard.toml` (Config as Code)

4. **Add Redis Service**:
   - Add Upstash Redis from Railway's marketplace
   - Or use your own Upstash Redis instance

5. **Set Environment Variables** (for each service):
   ```
   REDIS_URL=rediss://default:PASSWORD@HOST:6379
   PORT=8080  # (auto-set by Railway, but can override)
   ALLOWED_ORIGIN=https://your-frontend-url.vercel.app  # (Manager only)
   ```

6. **Configure Public Networking**:
   - **Manager**: Generate public domain on port `8080`
   - **Dashboard**: Generate public domain on port `8081` (for WebSocket)
   - **Worker**: No public domain needed (internal only)

### Frontend (Vercel)

1. **Deploy to Vercel**:
   ```bash
   # Install Vercel CLI
   npm i -g vercel
   
   # Deploy
   vercel
   ```

   Or connect your GitHub repo directly in Vercel dashboard.

2. **Set Environment Variables**:
   ```
   NEXT_PUBLIC_API_URL=https://your-manager-service.up.railway.app
   NEXT_PUBLIC_WS_URL=wss://your-dashboard-service.up.railway.app/ws
   ```

   **Important**: 
   - Use `https://` (not `http://`) for API URL
   - Use `wss://` (not `ws://`) for WebSocket URL
   - No trailing slashes in URLs

3. **Redeploy** after setting environment variables.

## 🔧 Environment Variables

### Frontend (Vercel)

| Variable | Description | Example |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_URL` | Manager service HTTP URL | `https://manager-production-xxx.up.railway.app` |
| `NEXT_PUBLIC_WS_URL` | Dashboard WebSocket URL | `wss://dashboard-production-xxx.up.railway.app/ws` |

### Backend Services (Railway)

| Variable | Service | Description | Example |
|----------|---------|-------------|---------|
| `REDIS_URL` | All | Redis connection string | `rediss://default:PASSWORD@HOST:6379` |
| `PORT` | All | Service port (auto-set) | `8080` |
| `ALLOWED_ORIGIN` | Manager | CORS allowed origin | `https://distri-task.vercel.app` |

## 📁 Project Structure

```
distritask/
├── app/                    # Next.js frontend
│   ├── dashboard/          # Dashboard page
│   │   └── page.tsx        # Main dashboard component
│   ├── layout.tsx          # Root layout
│   └── globals.css         # Global styles
├── cmd/                    # Go services
│   ├── manager/           # Manager service
│   │   └── main.go        # HTTP API for task submission
│   ├── worker/            # Worker service
│   │   └── main.go        # Task processor
│   └── dashboard/         # Dashboard WebSocket server
│       └── main.go        # WebSocket server
├── internal/              # Shared Go packages
│   ├── queue/            # Redis queue implementation
│   └── task/             # Task data structures
├── railway.*.toml        # Railway deployment configs
├── package.json          # Node.js dependencies
├── go.mod                # Go dependencies
└── README.md            # This file
```

## 🔄 How It Works

1. **Task Submission**:
   - User clicks "Deploy Node" in the frontend
   - Frontend sends `POST /submit` to Manager service
   - Manager creates a task and enqueues it to Redis Stream

2. **Task Processing**:
   - Worker service reads from Redis Stream
   - Worker processes the task (simulated work)
   - Worker broadcasts status updates via Redis Pub/Sub

3. **Real-time Updates**:
   - Dashboard WebSocket server subscribes to Redis Pub/Sub
   - When status updates are published, Dashboard server forwards them to connected WebSocket clients
   - Frontend receives updates and displays them in real-time

## 🛠️ Technology Stack

- **Frontend**: Next.js 15, React 19, TypeScript, Tailwind CSS, Framer Motion
- **Backend**: Go 1.25.5
- **Message Broker**: Redis (Upstash)
- **Deployment**: Railway (backend), Vercel (frontend)
- **WebSocket**: Gorilla WebSocket

## 🐛 Troubleshooting

### CORS Errors

- Ensure `ALLOWED_ORIGIN` in Railway Manager service matches your Vercel URL exactly
- No trailing slashes in URLs
- Use `https://` (not `http://`)

### 301 Redirect Errors

- Ensure `NEXT_PUBLIC_API_URL` has no trailing slash
- The code handles redirects automatically, but ensure URLs are correct

### Tasks Not Appearing

- Check Redis connection in all services
- Verify Worker service is running and connected to Redis
- Check Dashboard WebSocket server is running
- Verify WebSocket URL uses `wss://` (not `ws://`) in production

### Service Not Starting

- Check Railway logs for startup errors
- Verify `REDIS_URL` is set correctly (full connection string)
- Ensure `PORT` environment variable is set (Railway auto-sets this)

## 📝 License

MIT

## 🤝 Contributing

Contributions welcome! Please open an issue or submit a pull request.
