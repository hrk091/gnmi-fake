FROM golang:1.18-bullseye as build

RUN mkdir /src
WORKDIR /src
COPY go.mod ./
RUN go mod download

COPY . .

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64

RUN CGO_ENABLED=${CGO_ENABLED} \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    go build -o gnmi_fake -ldflags "-w -s" main.go


FROM scratch
WORKDIR /src
COPY --from=build /src/gnmi_fake /src/gnmi_fake

EXPOSE 8000

ENTRYPOINT ["/src/gnmi_fake"]
