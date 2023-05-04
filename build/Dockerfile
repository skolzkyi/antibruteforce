# Собираем в гошке
FROM golang:1.19 as build

ENV BIN_FILE /opt/antibruteforce/antibruteforce-app
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download github.com/sagikazarmark/crypt
RUN go mod download

COPY . ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/antibruteforce/*

# На выходе тонкий образ
FROM alpine:3.9

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="antibruteforce"
LABEL MAINTAINERS="skolzkyi@gmail.com"

ENV BIN_FILE "/opt/antibruteforce/antibruteforce-app"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/antibruteforce/config.env
COPY ./configs/dc/config.env ${CONFIG_FILE}

CMD ${BIN_FILE} -config ${CONFIG_FILE} > antibruteforceCLog.txt
