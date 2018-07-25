FROM alpine:3.7

WORKDIR /app

ADD . /app

# CMD ["/app/tnmd", "-inflators-file=/app/inflators.json"]
# CMD ["/app/tendermint", "node", "--home=/app/init"]
CMD ["/app/run.sh"]