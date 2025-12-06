# CAHIER DE CHARGES SIMPLIFIÉ - Système de Gestion de Bibliothèque

## 1. Contexte du Projet

Ce projet consiste à créer un système simple pour gérer une bibliothèque. Le système va permettre de gérer les livres, les utilisateurs et les emprunts. On va utiliser l'architecture SOA avec des services web REST et SOAP.

## 2. Objectifs du Projet

### Objectifs Principaux
- Créer un système qui permet de gérer les livres dans une bibliothèque
- Permettre aux utilisateurs de faire des emprunts de livres
- Gérer les retours des livres
- Avoir deux types de services : REST et SOAP
- Déployer avec Docker (optionnel)

### Objectifs Pédagogiques
- Comprendre l'architecture SOA
- Savoir développer des services REST et SOAP
- Gérer l'interopérabilité entre différents services
- Comprendre la communication inter-services

## 3. Périmètre Fonctionnel

### 3.1 Gestion des Livres (REST)
- Ajouter un nouveau livre
- Voir la liste de tous les livres
- Chercher un livre par titre
- Modifier les informations d'un livre
- Supprimer un livre
- Voir le nombre de copies disponibles

### 3.2 Gestion des Utilisateurs (REST)
- Créer un compte utilisateur
- Voir la liste des utilisateurs
- Voir un utilisateur par ID
- Modifier un utilisateur
- Supprimer un utilisateur

### 3.3 Gestion des Emprunts (SOAP)
- Emprunter un livre disponible
- Retourner un livre emprunté
- Voir les emprunts d'un utilisateur
- Voir tous les emprunts
- Voir un emprunt par ID

## 4. Architecture SOA Proposée

### 4.1 Services à Développer

**Service REST - Book Service** (Port 8081)
- Gestion complète des livres
- API REST moderne
- Format JSON pour les échanges
- CRUD complet (Create, Read, Update, Delete)

**Service REST - User Service** (Port 8082)
- Gestion des utilisateurs
- Format JSON
- CRUD complet

**Service SOAP - Loan Service** (Port 8083)
- Gestion des emprunts et retours
- Format XML pour les échanges
- Contrat WSDL bien défini
- Communication avec Book Service via REST

### 4.2 Communication entre Services
- Loan Service communique avec Book Service (REST) pour :
  - Vérifier si un livre existe
  - Vérifier si un livre est disponible
  - Mettre à jour la quantité disponible
- Loan Service communique avec User Service (REST) pour :
  - Vérifier si un utilisateur existe
- Utiliser HTTP pour toute la communication
- Gérer les erreurs de communication de base

## 5. Exigences Techniques Simplifiées

### 5.1 REST Services
- Utiliser les verbes HTTP correctement :
  - GET : récupérer des données
  - POST : créer des ressources
  - PUT : modifier des ressources
  - DELETE : supprimer des ressources
- Format JSON pour requests et responses
- Status codes HTTP appropriés (200, 201, 404, 400, etc.)

### 5.2 SOAP Service
- SOAP 1.2 avec WSDL
- Format XML pour les échanges
- Opérations bien définies
- Gestion des erreurs avec SOAP Faults

### 5.3 Base de Données
- Une seule base de données PostgreSQL partagée
- Trois tables : books, users, loans
- Relations simples avec foreign keys

## 6. Modèles de Données

### 6.1 Book
```
- id (auto-generated)
- isbn (unique)
- title
- author
- publishYear
- category
- availableQuantity
```

### 6.2 User
```
- id (auto-generated)
- username (unique)
- email
- firstName
- lastName
```

### 6.3 Loan
```
- id (auto-generated)
- userId (référence à User)
- bookId (référence à Book)
- loanDate
- dueDate (loanDate + 14 jours)
- returnDate (peut être null)
- status (ACTIVE ou RETURNED)
```

## 7. Livrables Attendus

### 7.1 Code Source
- Code des trois services (Book, User, Loan)
- Scripts de création de base de données
- Fichier de configuration (application.properties)
- README avec instructions de démarrage

### 7.2 Documentation
- Document expliquant l'architecture
- Liste des endpoints REST avec exemples
- Fichier WSDL pour le service SOAP
- Exemples de requêtes SOAP
- Instructions pour tester les services

## 8. Scénarios d'Utilisation Simplifiés

### Scénario 1 : Emprunter un Livre
1. Créer un utilisateur via User Service (REST POST)
2. Créer un livre via Book Service (REST POST)
3. Faire une demande d'emprunt via Loan Service (SOAP createLoan)
   - Loan Service vérifie si le livre existe
   - Loan Service vérifie si availableQuantity > 0
   - Loan Service crée l'emprunt
   - Loan Service décrémente availableQuantity du livre

### Scénario 2 : Retourner un Livre
1. User demande de retourner un livre via Loan Service (SOAP returnLoan)
2. Loan Service trouve l'emprunt
3. Loan Service met returnDate à aujourd'hui
4. Loan Service change status à RETURNED
5. Loan Service incrémente availableQuantity du livre

### Scénario 3 : Consulter les Livres
1. User demande la liste des livres via Book Service (REST GET /api/books)
2. Book Service retourne tous les livres avec leur disponibilité
3. User peut chercher un livre spécifique par titre (REST GET /api/books/search?title=...)

### Scénario 4 : Voir ses Emprunts
1. User demande ses emprunts via Loan Service (SOAP getLoansByUser)
2. Loan Service retourne la liste des emprunts (actifs et terminés)
