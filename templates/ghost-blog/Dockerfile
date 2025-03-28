FROM node:22.11.0 AS builder
WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

FROM node:22.11.0
WORKDIR /app
COPY --from=builder /app .
EXPOSE 3000

CMD ["npm", "run", "preview"]
