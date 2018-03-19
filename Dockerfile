FROM alpine

WORKDIR /app

ADD phoneserver phoneserver

# run our server
ENTRYPOINT ["./phoneserver"]

# expose port 8080
EXPOSE 8080