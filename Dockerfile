# Étape 1 : Builder (avec Go officiel, pas besoin d'alpine ici)
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compilation binaire statique
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gommunication main.go

# Étape 2 : Runtime ultra léger
FROM alpine:latest

# Ajout compatibilité glibc si besoin
RUN apk update && apk add --no-cache libc6-compat gcompat

# Création utilisateur non-root
RUN addgroup -g 1001 dns && adduser -D -u 1001 -G dns dns

# Copie du binaire depuis le builder
COPY --from=builder /app/gommunication /opt/gommunication
RUN chown dns:dns /opt/gommunication

USER dns
WORKDIR /opt
EXPOSE 8080
CMD ["./gommunication"]