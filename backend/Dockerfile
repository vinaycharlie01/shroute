FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

RUN microdnf install -y ca-certificates && microdnf clean all

ARG TARGETARCH=amd64
COPY backend/dist/linux_${TARGETARCH}/omnirouter /usr/local/bin/omnirouter

EXPOSE 8080

USER 1000:1000

ENTRYPOINT ["/usr/local/bin/omnirouter"]
