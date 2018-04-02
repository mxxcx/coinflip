FROM golang:latest

RUN mkdir -p /app

COPY ./db/migrations /app/db/migrations 

COPY ./build/linux/ns-game-api /app/

WORKDIR /app

EXPOSE 3000

CMD ["./ns-game-api"]