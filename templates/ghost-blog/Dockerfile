# Etapa 1: Construcción de la aplicación
FROM node:18.17.1-alpine as builder
WORKDIR /app

# Copiamos únicamente los package.json / package-lock.json
COPY package*.json ./
RUN npm install

# Copiamos el resto del proyecto y construimos
COPY . .
RUN npm run build

# Etapa 2: Servir la aplicación construida
FROM node:18.17.1-alpine
WORKDIR /app

# Copiamos los archivos construidos desde la etapa "builder"
COPY --from=builder /app .

# Exponemos el puerto en el que corre Astro (por defecto 3000)
EXPOSE 3000

# Usamos el script "preview" para levantar la app en modo producción
CMD ["npm", "run", "preview"]
