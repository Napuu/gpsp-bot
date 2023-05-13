FROM rust:1.67 as builder
WORKDIR /usr/src/myapp
COPY . .
RUN cargo install --path .

FROM ubuntu:rolling
RUN apt-get update && apt-get install -y ffmpeg yt-dlp libssl-dev && rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/local/cargo/bin/gpsp-bot /usr/local/bin/myapp
CMD ["myapp"]
