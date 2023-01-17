FROM golang:1.18 as build

WORKDIR /go/src/app

COPY . .

# build artifact in temp contaier
RUN go mod download
RUN go mod tidy
RUN GO111MODULE=on GOOS=linux go build -v -ldflags="-X 'main.GitCommitHash=$(git rev-parse HEAD)'" -o /main ./main.go
ADD . .

# copy artifacts to a clean image
FROM golang:1.18
COPY --from=build /main /main
RUN mkdir -p /config
COPY ./config /config
WORKDIR /
ENTRYPOINT [ "/main" ]


