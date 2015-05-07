#FROM scratch
#FROM gliderlabs/alpine:3.1
#FORM ubuntu_lean # https://blog.jtlebi.fr/2015/04/25/how-i-shrunk-a-docker-image-by-98-8-featuring-fanotify/

FROM ubuntu:14.04

#RUN apt-get update && apt-get install --no-install-recommends -y \
#    ca-certificates

COPY ./example /usr/bin/

EXPOSE 8000

CMD ["/usr/bin/example"]
