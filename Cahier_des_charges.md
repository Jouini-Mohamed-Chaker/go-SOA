# Cahier des Charges - API de Gestion de Bibliothèque

## 1. Présentation Générale

Ce système permet de gérer une bibliothèque avec des livres, des utilisateurs et des emprunts. Il est composé de 4 services qui communiquent entre eux.

## 2. Les 4 Services

### 2.1 Service d'Authentification (Port 8080)
**Rôle :** Sécuriser l'accès à l'application

**Points d'accès publics (sans connexion) :**
- **Inscription** : Créer un nouveau compte utilisateur
- **Connexion** : Obtenir un jeton de sécurité pour accéder aux autres services
- **Vérification** : Vérifier si un jeton est encore valide

**Points d'accès protégés (avec jeton) :**
- Accès aux livres, utilisateurs et emprunts
- Le jeton expire après 10 heures

### 2.2 Service des Livres (Port 8081)
**Rôle :** Gérer le catalogue de livres

**Fonctionnalités :**
- Voir tous les livres (avec pages)
- Chercher un livre par son titre
- Voir les détails d'un livre
- Ajouter un nouveau livre
- Modifier un livre existant
- Supprimer un livre

**Informations d'un livre :**
- Numéro ISBN
- Titre
- Auteur
- Année de publication
- Catégorie
- Quantité disponible

### 2.3 Service des Utilisateurs (Port 8082)
**Rôle :** Gérer les comptes utilisateurs

**Fonctionnalités :**
- Voir tous les utilisateurs (avec pages)
- Voir les détails d'un utilisateur
- Ajouter un utilisateur
- Modifier un utilisateur
- Supprimer un utilisateur

**Informations d'un utilisateur :**
- Nom d'utilisateur
- Email
- Prénom
- Nom de famille

### 2.4 Service des Emprunts (Port 8083)
**Rôle :** Gérer les emprunts de livres

**Fonctionnalités :**
- Créer un emprunt (durée : 14 jours automatiquement)
- Retourner un livre emprunté
- Voir les emprunts d'un utilisateur
- Voir tous les emprunts
- Voir un emprunt spécifique

**Informations d'un emprunt :**
- Identifiant de l'utilisateur
- Identifiant du livre
- Date d'emprunt
- Date de retour prévue
- Date de retour réelle
- Statut (ACTIF ou RETOURNÉ)

## 3. Règles de Fonctionnement

### 3.1 Sécurité
- Les services livres, utilisateurs et emprunts nécessitent un jeton de connexion
- Il faut d'abord s'inscrire, puis se connecter pour obtenir le jeton
- Le jeton doit être envoyé avec chaque demande protégée

### 3.2 Emprunts
- Un livre ne peut être emprunté que s'il est disponible (quantité > 0)
- Quand on emprunte un livre, la quantité disponible diminue de 1
- Quand on retourne un livre, la quantité disponible augmente de 1
- La durée d'emprunt est fixée à 14 jours

### 3.3 Livres
- Les champs obligatoires sont : ISBN, titre et auteur
- Les autres champs sont optionnels
- On peut chercher des livres par une partie du titre

### 3.4 Utilisateurs
- Le nom d'utilisateur et l'email doivent être uniques
- Les champs obligatoires sont : nom d'utilisateur et email
- Le prénom et nom de famille sont optionnels

## 4. Format des Données

- **Services Livres et Utilisateurs** : Format JSON (texte structuré)
- **Service Emprunts** : Format XML/SOAP (protocole ancien)
- **Service d'Authentification** : Format JSON

## 5. Pagination

Les listes de livres et d'utilisateurs utilisent un système de pages :
- Page par défaut : 1
- Nombre d'éléments par page par défaut : 10
- On peut changer ces valeurs dans la demande

## 6. Codes de Réponse

- **200** : Succès
- **201** : Création réussie
- **204** : Suppression réussie
- **400** : Erreur dans la demande
- **404** : Élément non trouvé

## 7. Exemples d'Utilisation

**Scénario typique :**
1. S'inscrire sur le service d'authentification
2. Se connecter et récupérer le jeton
3. Chercher un livre disponible
4. Créer un emprunt avec le jeton
5. Après lecture, retourner le livre

## 8. Technologies

- **Communication** : HTTP/REST pour la plupart des services, SOAP pour les emprunts
- **Format de données** : JSON et XML
- **Sécurité** : JWT (jeton web JSON)