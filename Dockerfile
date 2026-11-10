# ==============================================================================
# HOMOLOG - Plateforme Réglementaire ANPP & GED Pharmaceutique
# Dockerfile Multi-Stage de Production (Conformité BPF / Sécurité Souveraine)
# ==============================================================================

# Étape 1 : Compilation du binaire autonome Go
FROM golang:alpine AS builder

# Outils de build de base
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Téléchargement des dépendances (mise en cache des layers Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copie des sources complètes (y compris les assets web embarqués //go:embed)
COPY . .

# Compilation d'un binaire statique sans CGO (ultra-sécurisé, sans dépendance C externe)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X 'main.Version=0.7.0' -X 'main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
    -o /build/homolog ./cmd/server

# Étape 2 : Image d'exécution minimale et durcie (Minimal Hardened Runtime)
FROM alpine:3.20

# Installation des certificats TLS et de la base de fuseaux horaires
RUN apk add --no-cache ca-certificates tzdata curl

# Création d'un utilisateur non-root dédié pour l'exécution BPF
RUN addgroup -g 10001 homolog && \
    adduser -u 10001 -G homolog -s /bin/sh -D homolog

# Répertoires de données et de stockage CAS
RUN mkdir -p /var/lib/homolog/data/storage /var/log/homolog && \
    chown -R homolog:homolog /var/lib/homolog /var/log/homolog

WORKDIR /app

# Copie du binaire statique depuis l'étape de compilation
COPY --from=builder /build/homolog /app/homolog

# Permissions strictes
RUN chmod 755 /app/homolog

# Bascule sur l'utilisateur non privilégié (BPF Principle of Least Privilege)
USER homolog:homolog

# Variables d'environnement par défaut
ENV DEPLOYMENT_MODE=on-premise \
    PORT=8090 \
    STORAGE_PATH=/var/lib/homolog/data/storage \
    STORAGE_TYPE=local

# Volume persistant pour le stockage CAS chiffré
VOLUME ["/var/lib/homolog/data/storage"]

# Exposition du port applicatif
EXPOSE 8090

# Sonde de santé intégrée (Healthcheck Docker / Kubernetes)
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8090/health || exit 1

# Point d'entrée
ENTRYPOINT ["/app/homolog"]
