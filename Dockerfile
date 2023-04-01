FROM golang:1.20

RUN mkdir /app
COPY . /app
WORKDIR /app

EXPOSE 80

RUN ./golang_room_chat
CMD ["/app/main"]