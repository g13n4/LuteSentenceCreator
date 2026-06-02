FROM rust:1-alpine3.23

RUN apk update && apk add git

RUN mkdir -p /shared
WORKDIR /app

RUN git clone https://github.com/g13n4/sudachi.rs.git .
COPY system.dic /app/resources
COPY jpn_sentences.tsv /shared

RUN cargo build --release
RUN cargo install --path sudachi-cli/

RUN rm -f /shared/parsed.txt
RUN sudachi -w --split-sentences no /shared/jpn_sentences.tsv > /shared/parsed.txt
