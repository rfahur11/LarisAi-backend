# ============================================================
# LarisAI Backend - Hugging Face Spaces (Monolith Docker)
# Menggabungkan 3 microservices: POS (Go), AI Engine (Python),
# WhatsApp Service (Node.js) + Nginx reverse proxy di port 7860
# ============================================================

# ---- Stage 1: Build Go POS Service ----
FROM golang:1.21-alpine AS go-builder

WORKDIR /app/pos

RUN apk add --no-cache git ca-certificates tzdata

COPY pos/go.mod pos/go.sum ./
RUN go mod download

COPY pos/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o pos-server .


# ---- Stage 2: Final Runtime Image ----
FROM ubuntu:22.04

# Avoid prompts during apt installs
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Asia/Jakarta

# Install all runtimes: Python, Node.js, Nginx, Supervisor
RUN apt-get update && apt-get install -y \
    python3.11 \
    python3-pip \
    curl \
    nginx \
    supervisor \
    ca-certificates \
    tzdata \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# ---- Setup POS Go Service ----
WORKDIR /app/pos
COPY --from=go-builder /app/pos/pos-server .

# ---- Setup AI Engine (Python) ----
WORKDIR /app/ai-engine
COPY ai-engine/requirements.txt .
RUN pip3 install --no-cache-dir -r requirements.txt
COPY ai-engine/ .

# ---- Setup WhatsApp Service (Node.js) ----
WORKDIR /app/whatsapp
COPY whatsapp/package*.json ./
RUN npm install --production
COPY whatsapp/ .
RUN npm run build || true

# ---- Setup Nginx as Reverse Proxy ----
COPY hf-deploy/nginx.conf /etc/nginx/nginx.conf

# ---- Setup Supervisord to manage all processes ----
COPY hf-deploy/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Create log dirs
RUN mkdir -p /var/log/supervisor /app/whatsapp/auth_info_baileys

# Expose single port required by Hugging Face
EXPOSE 7860

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:7860/health || exit 1

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
