# Upframer Worker

Worker responsável pelo processamento de vídeos do sistema Upframer, desenvolvido em Go para extrair frames de vídeos e disponibilizá-los através de storage.

## 🏗️ Arquitetura

O projeto segue os princípios da **Clean Architecture**, com separação clara entre as camadas de domínio, aplicação e infraestrutura.

```
├── cmd/
│   └── consumer/           # Ponto de entrada da aplicação
├── internal/
│   ├── application/
│   │   └── usecases/      # Casos de uso da aplicação
│   ├── domain/
│   │   ├── entities/      # Entidades de domínio
│   │   ├── errors/        # Erros personalizados
│   │   ├── ports/         # Interfaces/contratos
│   │   └── services/      # Interfaces de serviços
│   └── infra/
│       ├── ffmpeg/        # Processador de vídeo (FFmpeg)
│       ├── rabbit/        # Cliente RabbitMQ
│       ├── storage/       # Adaptadores de storage (S3/Local)
│       └── util/          # Utilitários
```

## 🔧 Tecnologias

### Core
- **Go 1.22.5** - Linguagem principal
- **FFmpeg** - Processamento de vídeo e extração de frames

### Infraestrutura
- **RabbitMQ** - Sistema de mensageria
- **AWS S3** - Storage de arquivos (produção)
- **Docker** - Containerização
- **Alpine Linux** - Imagem base otimizada

### Dependências Principais
- `github.com/aws/aws-sdk-go-v2` - SDK AWS para Go
- `github.com/rabbitmq/amqp091-go` - Cliente RabbitMQ
- `github.com/joho/godotenv` - Carregamento de variáveis de ambiente

## 🚀 Funcionalidades

### Processamento de Vídeo
- Extração de frames de vídeos (1 frame por segundo)
- Suporte a vídeos locais e remotos (S3)
- Compactação dos frames em arquivo ZIP
- Upload automático para storage configurado

### Sistema de Filas
- Consumo de mensagens do RabbitMQ
- Sistema de retry com limite configurável (3 tentativas)
- Dead Letter Queue (DLQ) para mensagens com falha
- Classificação de erros (permanentes vs temporários)

### Storage Flexível
- **Produção**: AWS S3 obrigatório
- **Desenvolvimento**: S3 com fallback para storage local
- Download automático de vídeos do S3

## 🔄 Fluxo de Processamento

1. **Recebimento**: Worker consome mensagem da fila `job-creation`
2. **Download**: Se necessário, baixa o vídeo do S3
3. **Processamento**: Extrai frames usando FFmpeg (1 fps)
4. **Compactação**: Cria arquivo ZIP com os frames
5. **Upload**: Envia ZIP para storage configurado
6. **Notificação**: Publica resultado na fila `video-processing-result`
7. **Limpeza**: Remove arquivos temporários

## ⚙️ Configuração

### Variáveis de Ambiente

```bash
# Storage AWS S3
AWS_BUCKET=nome-do-bucket
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=sua-access-key
AWS_SECRET_ACCESS_KEY=sua-secret-key
AWS_SESSION_TOKEN=seu-session-token  # Opcional

# Ambiente
ENVIRONMENT=production|development

# Health Check
HEALTH_CHECK_PORT=3334
```

### Comportamento por Ambiente

#### Produção (`ENVIRONMENT=production`)
- AWS S3 **obrigatório**
- Falha fatal se credenciais S3 não fornecidas

#### Desenvolvimento
- AWS S3 **opcional**
- Fallback para storage local (`./output`)
- Logs informativos sobre storage utilizado

## 📊 Monitoramento

### Health Check
- **Endpoint**: `http://localhost:3334/health`
- **Intervalo**: 30s
- **Timeout**: 10s
- **Verificações**: Processo ativo + uso de disco < 90%

### Logs Estruturados
- Processamento de mensagens
- Falhas e retries
- Uso de storage (S3 vs Local)
- Limpeza de arquivos temporários

## Tratamento de Erros

### Classificação de Erros
- **Permanentes**: Arquivo não encontrado, formato inválido
- **Temporários**: Problemas de rede, storage indisponível

### Sistema de Retry
- **Máximo**: 3 tentativas
- **DLQ**: Mensagens com falha após máximo de retries
- **Headers**: Controle de contagem de retry

## 📦 Deploy

O projeto está configurado para deploy em **AWS ECS** com otimizações específicas:

- Limites de CPU e memória configurados
- Variables de ambiente para ECS
- Health checks compatíveis
- Logs estruturados para CloudWatch

## 🛠️ Desenvolvimento

### Pré-requisitos
- Go 1.22.5+
- FFmpeg instalado
- RabbitMQ em execução
- AWS CLI configurado (opcional)

### Executar Localmente
```bash
# Carregar dependências
go mod download

# Executar
go run cmd/consumer/main.go
```
