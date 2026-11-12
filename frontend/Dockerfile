FROM node:22-alpine
WORKDIR /app
COPY . .
RUN npm ci --legacy-peer-deps && npm run build
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
