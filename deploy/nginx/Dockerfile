FROM node:20-alpine AS frontend-builder

WORKDIR /web

COPY web/package*.json ./
RUN npm install

COPY web/ ./
ENV VUE_APP_API_BASE_URL=/api
RUN npm run build

FROM nginx:1.27-alpine

COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=frontend-builder /web/dist /usr/share/nginx/html

EXPOSE 80
