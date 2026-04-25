FROM node:20-alpine AS frontend-builder

WORKDIR /src/vue-frontend
COPY vue-frontend/package*.json ./
RUN npm ci
COPY vue-frontend/ ./
RUN npm run build

FROM nginx:1.27-alpine

COPY deploy/nginx/conf.d/gopherai.conf /etc/nginx/conf.d/default.conf
COPY --from=frontend-builder /src/vue-frontend/dist/ /usr/share/nginx/html/

