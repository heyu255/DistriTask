# Fix Railway Start Command Error

## The Problem
Railway is trying to run `./out` but the binary is named `dashboard`, `manager`, or `worker`.

## Solution: Update Start Command in Railway Dashboard

### For Dashboard Service:

1. Go to **Railway Dashboard** → **Dashboard Service** → **Settings**
2. Find **"Deploy"** or **"Start Command"** section
3. Set **Start Command** to: `./dashboard`
4. **Save** and **Redeploy**

### For Manager Service:

1. Set **Start Command** to: `./manager`
2. **Save** and **Redeploy**

### For Worker Service:

1. Set **Start Command** to: `./worker`
2. **Save** and **Redeploy**

## Why This Happens

Railway might be:
1. Using the `nixpacks.toml` start command instead of `railway.toml`
2. Using a cached start command
3. Not reading the `startCommand` from `railway.toml` correctly

## Alternative: Remove nixpacks.toml

If Railway keeps using `nixpacks.toml` instead of `railway.toml`:

1. **Rename** `nixpacks.toml` to `nixpacks.toml.backup`
2. **Push to GitHub**
3. **Redeploy** - Railway should now use `railway.toml` configs

## Verify

After setting the start command, check the logs. You should see:
- Dashboard: `./dashboard` starting
- Manager: `./manager` starting
- Worker: `./worker` starting

Instead of: `./out: No such file or directory`

