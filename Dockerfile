FROM golang:1.26-alpine as builder
WORKDIR /app

ARG APP_NAME
ENV APP_NAME ${APP_NAME}
ARG APP_VERSION
ENV APP_VERSION ${APP_VERSION}

COPY . .
RUN mkdir -p ./dist && GO111MODULE=on go mod download
RUN go build -ldflags "-X span.internal.Name=${APP_NAME} \
    -X span.internal.Version=${APP_VERSION}" \
    -o ./dist/span main.go

FROM alpine
WORKDIR /app

ARG APP_NAME
ENV APP_NAME ${APP_NAME}
ARG APP_VERSION
ENV APP_VERSION ${APP_VERSION}

# default config file
ENV CONF_FILE_PATH=/etc/span.yaml

COPY --from=builder /app/dist/span /app/span

CMD ["/app/span"]
