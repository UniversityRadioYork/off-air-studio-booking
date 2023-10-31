FROM golang:1.18

WORKDIR /usr/src/app

COPY . .
RUN go build

EXPOSE 8080

CMD [ "./off-air-studio-booking" ]