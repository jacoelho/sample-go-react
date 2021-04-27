FROM node:16-buster as assets
WORKDIR /sample
COPY client/package.json client/yarn.lock ./
RUN yarn install --frozen-lockfile
COPY client/. .
RUN yarn build


FROM golang:1.16-buster as build
WORKDIR /sample
COPY . .
COPY --from=assets /sample/build client/build
RUN make build


FROM gcr.io/distroless/base-debian10
COPY --from=build /go/bin/sample /
CMD ["/sample"]