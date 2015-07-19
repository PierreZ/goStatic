FROM centurylink/ca-certs

COPY goStatic /
ENTRYPOINT ["/goStatic"]