FROM node:26.10.0-bookworm-slim AS build
WORKDIR /app
COPY package.json ./
RUN npm install --ignore-scripts
COPY . .
ENV VITE_API_URL=""
RUN npm run build && test -f /app/dist/index.html

FROM nginxinc/nginx-unprivileged:1.31.5-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --chown=101:101 --from=build /app/dist/ /usr/share/nginx/html/
USER 101
EXPOSE 8080
