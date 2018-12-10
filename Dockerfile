FROM alpine:3.1
ADD ./node /root/node
CMD ["./root/node"]
