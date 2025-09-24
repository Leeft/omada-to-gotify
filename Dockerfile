FROM scratch
ADD ca-certificates.crt /etc/ssl/certs/
ADD omada-to-gotify /
EXPOSE 8080
CMD ["/omada-to-gotify"]