# Use official Node.js 18 LTS
FROM node:18-alpine

# Create app dir
WORKDIR /usr/src/app

# Copy package manifests and install deps
COPY package*.json ./
RUN npm install --production

# Copy source
COPY . .

# Expose nothing (we're a consumer), but define default command
CMD ["node", "app.js"]