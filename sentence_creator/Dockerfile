FROM golang:1.26.3-bookworm

RUN mkdir -p /shared

WORKDIR /mhs

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    g++ \
    git \
    make \
    libboost-system-dev \
    libboost-program-options-dev \
    libboost-log-dev \
    && rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/g13n4/Minimal-Hitting-Set-Algorithms .

RUN make

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /target .

EXPOSE 9999

CMD ["/target"]
