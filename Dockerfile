FROM docker.io/rust:1.68 as bot-builder
WORKDIR /usr/src/myapp
COPY . .
ENV CARGO_REGISTRIES_CRATES_IO_PROTOCOL=sparse
RUN cargo build --release --quiet

FROM docker.io/ubuntu:lunar as util-builder
RUN apt-get update && apt-get install build-essential git zip python3 -y > /dev/null 2>&1
RUN cd /tmp && git clone https://github.com/yt-dlp/yt-dlp --depth=1
RUN cd /tmp/yt-dlp && make yt-dlp

FROM docker.io/ubuntu:lunar
RUN apt-get update && apt-get install python3 ffmpeg -y > /dev/null 2>&1 && apt-get clean
COPY --from=bot-builder /usr/src/myapp/target/release/gpsp-bot /usr/local/bin/myapp
COPY --from=util-builder /tmp/yt-dlp/yt-dlp /usr/local/bin/yt-dlp

ENTRYPOINT ["/usr/local/bin/myapp"]
