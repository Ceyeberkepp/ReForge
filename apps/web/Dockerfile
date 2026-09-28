FROM node:24.11.0-bookworm-slim AS build
WORKDIR /app
COPY package.json ./
RUN npm install --ignore-scripts
COPY . .
ENV VITE_API_URL=""
RUN npm run build

FROM nginxinc/nginx-unprivileged:1.31.5-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
USER 101
EXPOSE 8080
