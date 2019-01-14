FROM alpine:3.1
ADD ./nctler /root/nctler
CMD ["./root/nctler"]
