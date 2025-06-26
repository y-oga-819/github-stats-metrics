FROM node:22-bullseye

ENV ROOT /home/node/app
WORKDIR ${ROOT}

# SSL証明書問題対応
RUN npm config set strict-ssl false