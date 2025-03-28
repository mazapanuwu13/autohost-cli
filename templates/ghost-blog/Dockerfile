# Etapa 1: Construcción de la aplicación
FROM node:16-alpine as builder
WORKDIR /app

# Copiar archivos de definición e instalar dependencias
COPY package*.json ./
RUN npm install

# Copiar el resto del código fuente y construir la app
COPY . .
RUN npm run build

# Etapa 2: Servir la aplicación
FROM node:16-alpine
WORKDIR /app

# Copiar los archivos construidos desde la etapa builder
COPY --from=builder /app ./

# Exponer el puerto en el que corre Astro (por defecto 3000)
EXPOSE 3000

# Comando para iniciar la aplicación
CMD ["npm", "start"]
