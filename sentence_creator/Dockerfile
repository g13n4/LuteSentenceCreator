# syntax=docker/dockerfile:1
FROM golang:1.26.3-bookworm AS builder

# Setting up minimla hitting set first
WORKDIR /mhs
RUN clone https://github.com/g13n4/Minimal-Hitting-Set-Algorithms

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    g++ \
    make \
    libboost-system-dev \
    libboost-program-options-dev \
    libboost-log-dev \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /execs

RUN make clean && make -j$(nproc)
RUN ln -s /agdmhs /execs/

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /target .

EXPOSE 9999

CMD ["/target"]
