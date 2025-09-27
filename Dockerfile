# Multi-stage build otimizado para ECS
FROM golang:1.22-alpine AS builder

# Instalar dependências de build
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências primeiro (melhor cache de layers)
COPY go.mod go.sum ./

# Download das dependências
RUN go mod download && go mod verify

# Copiar apenas arquivos necessários para o build
COPY cmd/ cmd/
COPY internal/ internal/

# Build da aplicação com otimizações para produção
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o consumer ./cmd/consumer/main.go

# Stage 2: Imagem final otimizada para ECS
FROM alpine:3.19

# Instalar dependências de runtime otimizadas para ffmpeg
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    ffmpeg \
    ffmpeg-dev \
    && rm -rf /var/cache/apk/* /tmp/* /var/tmp/*

# Criar usuário não-root para segurança
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Definir diretório de trabalho
WORKDIR /app

# Copiar binário da aplicação do stage anterior
COPY --from=builder /app/consumer .

# Criar diretórios necessários com permissões corretas
RUN mkdir -p /app/output /app/frames /app/temp && \
    chown -R appuser:appgroup /app

# Configurações específicas para ECS Task
# Limites de memória e CPU otimizados para processamento de vídeo
ENV GOMAXPROCS=2
ENV GOMEMLIMIT=1GiB

# Configurações de ffmpeg otimizadas para containers
ENV FFMPEG_THREADS=2
ENV FFMPEG_PRESET=fast
ENV FFMPEG_LOG_LEVEL=error

# Configurações de ambiente para ECS
ENV ENVIRONMENT=production
ENV AWS_REGION=us-east-1

# Health check otimizado para ECS com verificação de recursos
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD pgrep consumer > /dev/null && \
        df /app | awk 'NR==2{if($5+0 < 90) exit 0; else exit 1}' || exit 1

# Trocar para usuário não-root
USER appuser

# Comando padrão
CMD ["./consumer"]