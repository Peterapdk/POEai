# Running OpenClaw in Termux proot-distro debian

## Overview
This is a quick sharing of the catches when installing OpenClaw in proot debian.

**Important Notes:**
- Alpine won't work because @mariozechner/clipboard native plugin does not have linux arm64 musl version
- Device Used: Asus ROG phone 9 pro edition (SD 8 Elite, 24GB RAM)
- Installation is done in a separate App running proot using the Termux proot binary

## Limitations
Unfortunately, running a local model on the phone is not feasible. Running a model as an agent requires 10x the power compared to chatting. Currently only using Gemini API for this guide.

---

## Dependencies

### Recommended Tools
- **fish** shell: Better auto-completions and command validation
- **micro** editor: For viewing files

### DNS Configuration
Make sure you can run `apt update` properly. If not, use:

```bash
printf "nameserver 8.8.8.8\nnameserver 1.1.1.1">/etc/resolv.conf
```

### Install Required Packages

```bash
apt install curl git cmake build-essential zstd unzip
```

---

## Installing Node.js and Bun

**Note:** Install Node.js first, then use Bun to run OpenClaw. Using Node.js directly is too slow, even for `openclaw --help`.

### Node.js Installation

1. Use the official script from the Node.js website
2. Save it using: `micro inode.sh` then `sh inode.sh`
3. Add Node.js to your PATH on every login:

```bash
export PATH="/root/versions/node/v25.5.0/bin:$PATH"
```

### Bun Installation

1. Use the official Bun install script:

```bash
curl -fsSL https://bun.sh/install | bash
```

2. Export to PATH manually on every login:

```bash
export BUN_INSTALL="$HOME/.bun"
export PATH="${BUN_INSTALL}/bin:$PATH"
```

---

## Installing and Patching OpenClaw

### npm Installation

```bash
npm i -g openclaw
```

**Note:** This builds llama.cpp in the background and takes a very long time.

### Required Patches

After installing, do NOT run openclaw directly. Some patches are needed for proper operation.

#### Patch os.networkInterfaces()

The os.networkInterfaces() function needs a static value. Follow these steps:

1. **Generate the static value** (in native Termux):

```bash
cd $PREFIX/var/lib/proot-distro/installed-rootfs/debian/root
node -e 'fs.writeFileSync("./ni.json",JSON.stringify(os.networkInterfaces()))'
```

2. **Files to patch** under `~/versions/node/v25.5.0/lib/node_modules/`:
   - `openclaw/dist/infra/system-presence.js`
   - `openclaw/dist/infra/tailnet.js`
   - `openclaw/node_modules/@homebridge/ciao/lib/NetworkManager.js`

3. **For the first two files:**
   - Add at the beginning:
     ```javascript
     import fs from "node:fs";
     ```
   - Search for `os.networkInterfaces()` and replace with:
     ```javascript
     JSON.parse(fs.readFileSync('/root/ni.json').toString())
     ```

4. **For the third file:**
   - Search for `os_1.default.networkInterfaces()` and replace with:
     ```javascript
     JSON.parse(require('fs').readFileSync('/root/ni.json').toString())
     ```

---

## Running OpenClaw & Authenticating with Gemini CLI

### Install Gemini CLI

```bash
npm i -g @google/gemini-cli
```

### Ensure xdg-open Works

Make sure xdg-open can open a browser window for authentication. Options:

**Option 1: Install a GUI** (recommended)

**Option 2: Use a workaround:**

```bash
mv /usr/bin/xdg-open /usr/bin/xdg-open.bak

printf '#!/bin/sh
echo "$@">>/data/data/com.termux/files/home/authurl.txt'>/usr/bin/xdg-open

chmod +x /usr/bin/xdg-open
```

Remember to restore xdg-open after authentication if needed.

### Run Onboard Setup

```bash
bun $(which openclaw) onboard
```

This will guide you through setting up Gemini API as the provider.

If you used the xdg-open workaround, open a separate Termux tab and check the auth URL:

```bash
cat authurl.txt
```

### Get the Access Token

In proot debian:

```bash
micro ~/.openclaw/openclaw.json
```

Look for `gateway.auth.token` and copy it.

---

## Running the Gateway (WebUI Server)

```bash
bun $(which openclaw) gateway &
```

Open in your browser: `http://localhost:18789`

Paste the auth token from the previous section and start your journey.

### Security Recommendation

For isolated operation, use:

```bash
pd login --isolated debian
```

Before actually running the gateway.

### Optional: Graphical Apps Support

If you want graphical apps to work, run in Termux native:

```bash
termux-x11 :0 -listen tcp -ac &
export DISPLAY=:0
LIBGL_ALWAYS_SOFTWARE=1 xfwm4 &
```

Then run X11 over TCP in proot:

```bash
export DISPLAY=:0
konsole
# or some other app
```

---

## Browser Automation

**Reference:** https://youtube.com/shorts/EsimoncQJw8?si=d-NA0Rkl5nqPttYy

### Important Notes

- This function requires Node.js (NOT Bun) for the gateway
- Additional patch needed for `openclaw/dist/pw-ai-*.js`

### Setup Steps

1. **Hardcode the endpoint** in `openclaw/dist/pw-ai-*.js`:

```javascript
const endpoint = "http://127.0.0.1:18791"; //
```

2. **Configure OpenClaw** in the config -> browser section:
   - Turn on "No Sandbox"
   - Turn on "Browser Evaluate Enabled"
   - Set the default profile to "openclaw"
   - Set CDP URL to: `http://127.0.0.1:18791`

3. **Test the setup:**

```bash
# Open a test browser window
openclaw browser --browser-profile=openclaw open https://reddit.com

# Test the automation function
openclaw browser --browser-profile=openclaw snapshot --efficient
```

---

## Summary

This guide covers the complete setup for running OpenClaw in Termux with proot-distro debian. The main challenges are:

1. Patching os.networkInterfaces() for proper network detection
2. Using Bun instead of Node.js for performance
3. Configuring Gemini API authentication
4. Optional browser automation setup

For the best experience, use Gemini API as your AI provider and Bun as your runtime.