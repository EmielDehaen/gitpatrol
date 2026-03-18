# Frontend Dockerfile (Development Mode)
FROM node:22.12-alpine AS dev
WORKDIR /app

# Install Bun
RUN npm install -g bun

# Copy package files
COPY package.json bun.lock ./
RUN bun install

# Copy source
COPY . .

# Vite default port
EXPOSE 5173

# Run Vite dev server with host flag to allow access from outside the container
CMD ["bun", "run", "dev", "--host"]
