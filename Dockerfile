# Сборка: swagger.json генерируется прямо в образе (go generate запускает swag),
# в репозитории его нет. Итоговый образ без Go, один статический бинарник

FROM golang:1.26.8-alpine AS build

# в закрытом контуре модули берутся с внутреннего прокси:
# docker build --build-arg GOPROXY=https://<прокси> .
ARG GOPROXY

WORKDIR /src

# версия swag совпадает с README
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# timetzdata вшивает базу часовых поясов в бинарник: окна тасок считаются во
# времени сервера, и TZ=Europe/Moscow должен работать без tzdata в образе
RUN go generate ./... \
 && CGO_ENABLED=0 go build -trimpath -tags timetzdata -ldflags="-s -w" -o /out/footprints .


FROM alpine:3.21

COPY --from=build /out/footprints /footprints

# пишем только в базу и в сеть, корневая файловая система может быть только на чтение
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/footprints"]
