FROM golang:1.23.0-bookworm AS build

ENV CGO_ENABLED=1

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG BUILD=dev

WORKDIR /build
COPY . .
RUN apt update && apt install libgdbm-dev
RUN git clone https://git.defalsify.org/vise.git go-vise

WORKDIR /build/services/registration
RUN echo "Compiling go-vise files"
RUN make VISE_PATH=/build/go-vise -B

WORKDIR /build
RUN echo "Building on $BUILDPLATFORM, building for $TARGETPLATFORM"
RUN go mod download
RUN go build -tags logtrace -o ussd-africastalking -ldflags="-X main.build=${BUILD} -s -w" cmd/africastalking/main.go
RUN go build -tags logtrace -o ussd-ssh -ldflags="-X main.build=${BUILD} -s -w" cmd/ssh/main.go

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && apt install libgdbm-dev ca-certificates -y
RUN apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /service

COPY --from=build /build/ussd-africastalking .
COPY --from=build /build/ussd-ssh .
COPY --from=build /build/LICENSE .
COPY --from=build /build/README.md .
COPY --from=build /build/services ./services
COPY --from=build /build/.env.example .
RUN mv .env.example .env

EXPOSE 7123
EXPOSE 7122

CMD ["./ussd-africastalking"]