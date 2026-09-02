FROM docker.io/golang:1.27.1
RUN apt-get update && \
  apt-get install ffmpeg yt-dlp tesseract-ocr -y && \
  go run github.com/mxschmitt/playwright-go/cmd/playwright@v0.6201.1 install --with-deps chromium && \
  apt-get clean

WORKDIR /app
COPY . .

RUN go build .
ENTRYPOINT ["./gpsp-bot"]