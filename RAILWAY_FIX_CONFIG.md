# Fix Railway Still Using Default Build Command

## The Problem
Railway is still using `go build -o distritask ./cmd/manager` even after selecting the config file.

## Why This Happens
Railway might be:
1. **Prioritizing `nixpacks.toml`** over `railway.toml`
2. **Using cached build settings**
3. **Not detecting the config file correctly**

## Solution: Multiple Approaches

### Option 1: Remove/Disable nixpacks.toml (Recommended)

The `nixpacks.toml` file might be overriding your `railway.toml` config. Try:

1. **Temporarily rename** `nixpacks.toml` to `nixpacks.toml.backup`
2. **Push to GitHub**
3. **Redeploy** in Railway
4. Railway should now use `railway.dashboard.toml`

### Option 2: Update nixpacks.toml to Match

Instead of removing it, update `nixpacks.toml` to use the same build command as your railway configs, but this won't work for multiple services.

### Option 3: Use Railway Dashboard Settings Directly

If Config as Code isn't working:

1. Go to Railway Dashboard → Dashboard Service → **Settings**
2. Find **"Build"** or **"Deploy"** section
3. **Manually set**:
   - **Build Command**: `go build -o dashboard cmd/dashboard/main.go`
   - **Start Command**: `./dashboard`
4. **Save** and **Redeploy**

### Option 4: Clear Build Cache

Railway might be using cached build settings:

1. Go to Railway Dashboard → Dashboard Service → **Settings**
2. Look for **"Clear Cache"** or **"Rebuild"** option
3. Or add environment variable: `NO_CACHE=1`
4. **Redeploy**

## Recommended Steps (Try in Order)

1. **Verify Config File is Selected**:
   - Railway Dashboard → Dashboard Service → Settings → Config as Code
   - Make sure `railway.dashboard.toml` is selected
   - Save

2. **Temporarily Rename nixpacks.toml**:
   ```bash
   # Rename it so Railway doesn't use it
   mv nixpacks.toml nixpacks.toml.backup
   ```
   Push and redeploy

3. **If Still Not Working**:
   - Delete the Dashboard service
   - Create new service
   - **Before first deploy**, select `railway.dashboard.toml` in Config as Code
   - Then deploy

## Verify It's Working

After redeploying, check the build logs. You should see:
```
go build -o dashboard cmd/dashboard/main.go
```

NOT:
```
go build -o distritask ./cmd/manager
```

