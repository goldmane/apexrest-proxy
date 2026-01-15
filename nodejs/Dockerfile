FROM node:20-alpine

WORKDIR /app

COPY package.json .
COPY server.js .

RUN npm install --omit=dev

ENV PORT=8080
EXPOSE $PORT

CMD ["npm", "start"]