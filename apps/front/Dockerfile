# syntax=docker/dockerfile:1

FROM oven/bun:1-slim AS deps
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

FROM oven/bun:1-slim AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1

# NEXT_PUBLIC_* vars are inlined into the client bundle at build time, so they
# must be passed as --build-arg, not as runtime container env vars.
ARG NEXT_PUBLIC_PREREGISTRATION
ARG NEXT_PUBLIC_ACTIVE_SEASON_ID
ARG NEXT_PUBLIC_DOMAIN
ARG NEXT_PUBLIC_BACKEND_URL
ARG NEXT_PUBLIC_ORY_SDK_URL
ARG NEXT_PUBLIC_SUPPORT_EMAIL
ARG NEXT_PUBLIC_SUPPORT_URL
ARG NEXT_PUBLIC_TELEGRAM_CHANNEL_URL
ARG NEXT_PUBLIC_DISCORD_SERVER_URL
ENV NEXT_PUBLIC_PREREGISTRATION=${NEXT_PUBLIC_PREREGISTRATION} \
    NEXT_PUBLIC_IMGPROXY_ENDPOINT=${NEXT_PUBLIC_IMGPROXY_ENDPOINT} \
    NEXT_PUBLIC_ACTIVE_SEASON_ID=${NEXT_PUBLIC_ACTIVE_SEASON_ID} \
    NEXT_PUBLIC_DOMAIN=${NEXT_PUBLIC_DOMAIN} \
    NEXT_PUBLIC_BACKEND_URL=${NEXT_PUBLIC_BACKEND_URL} \
    NEXT_PUBLIC_ORY_SDK_URL=${NEXT_PUBLIC_ORY_SDK_URL} \
    NEXT_PUBLIC_SUPPORT_EMAIL=${NEXT_PUBLIC_SUPPORT_EMAIL} \
    NEXT_PUBLIC_SUPPORT_URL=${NEXT_PUBLIC_SUPPORT_URL} \
    NEXT_PUBLIC_TELEGRAM_CHANNEL_URL=${NEXT_PUBLIC_TELEGRAM_CHANNEL_URL} \
    NEXT_PUBLIC_DISCORD_SERVER_URL=${NEXT_PUBLIC_DISCORD_SERVER_URL}

RUN bun run build

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs \
  && adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

CMD ["node", "server.js"]
