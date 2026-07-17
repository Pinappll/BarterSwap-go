# BarterSwap-go — API d'échange de compétences

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

```bash
go run .
```

L'API est accessible sur http://localhost:8080.

## Endpoints

### Utilisateurs

| Méthode | Path                      | Auth | Description                                   |
| ------- | ------------------------- | :--: | --------------------------------------------- |
| POST    | `/api/users`              |      | Créer un compte (10 crédits de bienvenue)     |
| GET     | `/api/users/{id}`         |      | Profil public d'un utilisateur                |
| PUT     | `/api/users/{id}`         |  ✅  | Modifier son profil (propriétaire uniquement) |
| GET     | `/api/users/{id}/skills`  |      | Compétences d'un utilisateur                  |
| PUT     | `/api/users/{id}/skills`  |  ✅  | Définir ses compétences (écrase la liste)     |
| GET     | `/api/users/{id}/stats`   |      | Statistiques d'un utilisateur                 |
| GET     | `/api/users/{id}/reviews` |      | Avis reçus par un utilisateur                 |

### Services

| Méthode | Path                         | Auth | Description                                                        |
| ------- | ---------------------------- | :--: | ------------------------------------------------------------------ |
| GET     | `/api/services`              |      | Liste des services actifs (filtres `categorie`, `ville`, `search`) |
| POST    | `/api/services`              |  ✅  | Créer une annonce (compétence requise)                             |
| GET     | `/api/services/{id}`         |      | Détail d'un service                                                |
| PUT     | `/api/services/{id}`         |  ✅  | Modifier son annonce (propriétaire uniquement)                     |
| DELETE  | `/api/services/{id}`         |  ✅  | Supprimer son annonce (409 si un échange est en cours)             |
| GET     | `/api/services/{id}/reviews` |      | Avis reçus sur un service                                          |

### Échanges

| Méthode | Path                           | Auth | Description                                                          |
| ------- | ------------------------------ | :--: | -------------------------------------------------------------------- |
| POST    | `/api/exchanges`               |  ✅  | Créer une demande d'échange                                          |
| GET     | `/api/exchanges`               |  ✅  | Mes échanges (requêtes + reçus, filtre `status`)                     |
| GET     | `/api/exchanges/{id}`          |  ✅  | Détail d'un échange (participants uniquement)                        |
| PUT     | `/api/exchanges/{id}/accept`   |  ✅  | Accepter (offreur uniquement) — bloque les crédits                   |
| PUT     | `/api/exchanges/{id}/reject`   |  ✅  | Refuser (offreur uniquement)                                         |
| PUT     | `/api/exchanges/{id}/complete` |  ✅  | Marquer comme terminé (demandeur uniquement) — transfère les crédits |
| PUT     | `/api/exchanges/{id}/cancel`   |  ✅  | Annuler (demandeur ou offreur) — rembourse si déjà accepté           |
| POST    | `/api/exchanges/{id}/review`   |  ✅  | Donner un avis sur un échange terminé                                |

## Exemples d'utilisation

### 1. Créer un compte et consulter son profil

```bash
curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"pseudo": "alice", "ville": "Nantes"}'

curl -s http://localhost:8080/api/users/1
```

### 2. Déclarer une compétence puis publier un service

```bash
curl -s -X PUT http://localhost:8080/api/users/1/skills \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '[{"nom": "Jardinage", "niveau": "expert"}]'

curl -s -X POST http://localhost:8080/api/services \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{"titre": "Taille de haie", "categorie": "Jardinage", "duree_minutes": 60, "credits": 3, "ville": "Nantes"}'
```

### 3. Cycle de vie complet d'un échange

```bash
curl -s -X POST http://localhost:8080/api/exchanges \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 2" \
  -d '{"service_id": 1}'

curl -s -X PUT http://localhost:8080/api/exchanges/1/accept -H "X-User-ID: 1"

curl -s -X PUT http://localhost:8080/api/exchanges/1/complete -H "X-User-ID: 2"

curl -s http://localhost:8080/api/users/1 | grep -o '"credit_balance":[0-9]*'   # 13
curl -s http://localhost:8080/api/users/2 | grep -o '"credit_balance":[0-9]*'   # 7
```

### 4. Laisser un avis puis consulter les statistiques

```bash
curl -s -X POST http://localhost:8080/api/exchanges/1/review \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 2" \
  -d '{"note": 5, "commentaire": "Travail impeccable"}'

curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8080/api/exchanges/1/review \
  -H "Content-Type: application/json" -H "X-User-ID: 2" -d '{"note": 3}'

curl -s http://localhost:8080/api/users/1/stats
```

## Tests

```bash
go test -v -cover ./...
```
