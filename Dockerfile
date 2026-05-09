# Distroless final image; binary is built by GoReleaser and copied in.
FROM gcr.io/distroless/static-debian12:nonroot
COPY logz /usr/local/bin/logz
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/logz"]
