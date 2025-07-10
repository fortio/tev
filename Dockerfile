FROM scratch
COPY tev /usr/bin/tev
ENTRYPOINT ["/usr/bin/tev"]
