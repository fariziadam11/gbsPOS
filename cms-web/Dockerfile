# Multi-stage build for the GBS CMS web frontend
# Stage 1: Build the Vite + Vue SPA with Bun
FROM oven/bun:1.3.14 AS builder

WORKDIR /app

# Copy package manifest first for better layer caching
COPY package.json ./
RUN bun install

# Copy source and build
COPY . .
RUN bunx vite build

# Stage 2: Serve static files with nginx
FROM nginx:1.27-alpine

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy SPA nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
