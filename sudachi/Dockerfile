FROM rust:1-alpine3.23

RUN apk update && apk add git

RUN mkdir -p /shared

WORKDIR /app

COPY *.dic *.tsv /tmp/

RUN git clone https://github.com/g13n4/sudachi.rs.git .
RUN  mv /tmp/*.dic /app/resources/system.dic

RUN cargo build --release
RUN cargo install --path sudachi-cli/

RUN  mv /tmp/*.tsv /shared/jpn_sentences.tsv

RUN sudachi -w --split-sentences no /shared/jpn_sentences.tsv > /shared/parsed.txt
