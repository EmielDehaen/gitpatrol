# Frontend Dockerfile (Development Mode)
FROM oven/bun:1.1 AS dev
WORKDIR /app

# Copy package files
COPY package.json bun.lock ./
RUN bun install

# Copy source
COPY . .

# Vite default port
EXPOSE 5173

# Run Vite dev server with host flag to allow access from outside the container
CMD ["bun", "run", "dev", "--host"]
