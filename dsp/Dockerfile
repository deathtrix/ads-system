FROM golang:1.14-alpine

WORKDIR ./dsp
COPY . .


RUN apk update && apk upgrade && \
    apk add --no-cache bash git

RUN go get -u .
# RUN go install -v .
RUN go build -o main .

CMD ["./main"]
