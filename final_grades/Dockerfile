# Dockerfile
FROM node:20-alpine

# 1. Create app directory
WORKDIR /usr/src/app

# 2. Install dependencies
COPY package.json package-lock.json ./
RUN npm ci --production

# 3. Copy source
COPY . .

# 4. Expose nothing (consumer only), set default ENV
ENV NODE_ENV=production

# 5. Default command: run consumer
CMD ["node", "grades_consumer.js"]
