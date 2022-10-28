FROM alpine:latest

RUN mkdir /app

COPY userBalanceApp /app

CMD [ "/app/userBalanceApp"]