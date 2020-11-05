FROM golang:1.14 AS builder
WORKDIR /fountain
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . ./
RUN make build

FROM alpine:3.12
ENV FOUNTAIN_DESTINATION /s3

COPY --from=builder /fountain/fountain /usr/local/bin/
RUN mkdir /s3
CMD ["fountain"]
