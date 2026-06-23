# BarterSwap-go — API d'échange de compétences

Ce projet est une API REST développée en Go (stdlib uniquement) avec une base de données PostgreSQL. L'environnement de développement est entièrement conteneurisé pour garantir la même exécution sur toutes les machines.

## 🛠 Prérequis

- **Git**
- **Docker** et le plugin **Docker Compose**

## Installation & Démarrage rapide

### 1. Cloner le dépôt

```bash
git clone git@github.com:Pinappll/BarterSwap-go.git
cd barterswap
```

### 2. Configurer l'environnement

Copiez le fichier d'exemple des variables d'environnement :

```bash
cp .env.example .env
```

### 3. Lancer les conteneurs (Base de données + Environnement Go)

```bash
docker compose up -d --build
```

### 4. Initialiser la base de données

```bash
docker compose exec -T db psql -U postgres -d barterswap < schema.sql
```

## Développement

### 1. Entrez dans le conteneur Go

```bash
docker compose exec go bash
```

### 2. Lancez le serveur HTTP

```go
go run .
```

L'API est accessible sur http://localhost:8080.
