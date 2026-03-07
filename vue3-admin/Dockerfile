FROM node:18-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates && update-ca-certificates
RUN npm config set registry https://registry.npmmirror.com

COPY package.json ./
RUN npm install --legacy-peer-deps

COPY . .

RUN npm run build:release

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html

COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
