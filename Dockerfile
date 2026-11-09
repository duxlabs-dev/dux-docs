FROM node:22-alpine AS deps
RUN apk add --no-cache libc6-compat python3 make g++
WORKDIR /app

COPY package*.json ./
COPY patches ./patches

RUN npm ci

FROM node:22-alpine AS builder
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Copy dependencies from deps stage
COPY --from=deps /app/node_modules ./node_modules

# Copy source files
COPY . .

ARG NODE_ENV=production
ENV NODE_ENV=${NODE_ENV}

RUN npm run build

FROM node:22-alpine AS production-node
RUN apk add --no-cache libc6-compat
WORKDIR /app

RUN npm install -g serve@14.2.7 && \
    node -e "const fs=require('fs'),f=process.argv[1]+'/serve/node_modules/serve-handler/src/index.js',a='stats = await handlers.lstat(absolutePath);',s=fs.readFileSync(f,'utf8');if(!s.includes(a))process.exit(1);fs.writeFileSync(f,s.replace(a,a+' if (stats.isDirectory()) stats = null;'))" "$(npm root -g)"

COPY --from=builder /app/build ./build

RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

RUN chown -R nodejs:nodejs /app

USER nodejs

EXPOSE 3000

CMD ["serve", "build", "-l", "3000"]
