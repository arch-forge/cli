FROM alpine:3.21

RUN apk add --no-cache ca-certificates git

COPY arch_forge /usr/local/bin/arch_forge

RUN addgroup -S archforge && adduser -S -G archforge archforge
USER archforge

WORKDIR /workspace

ENTRYPOINT ["arch_forge"]
CMD ["--help"]
