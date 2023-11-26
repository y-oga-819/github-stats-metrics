FROM node:21.2.0-bullseye

ENV ROOT /home/node/app
WORKDIR ${ROOT}

COPY package.json ./
RUN npm install

CMD ["yarn", "dev"]