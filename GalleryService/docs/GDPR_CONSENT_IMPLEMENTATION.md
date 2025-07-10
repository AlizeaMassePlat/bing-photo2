# Implémentation du Système de Consentement RGPD

## Vue d'ensemble

Cette implémentation ajoute un système complet de gestion des consentements RGPD à l'application de galerie photo, avec un focus particulier sur la fonctionnalité de détection d'images similaires.

## Architecture

### Modèles de données

#### Table `consents`
```sql
CREATE TABLE consents (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    consent_type VARCHAR(255) NOT NULL,
    consent_version VARCHAR(50) NOT NULL,
    is_granted BOOLEAN NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    consent_text TEXT,
    revoked_at TIMESTAMP,
    revoked_ip_address VARCHAR(45),
    revoked_user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Services

#### ConsentService
- `AddConsent()` : Ajoute un nouveau consentement
- `GetUserConsents()` : Récupère tous les consentements d'un utilisateur
- `GetActiveConsent()` : Récupère le consentement actif pour un type
- `RevokeConsent()` : Révoque un consentement
- `HasValidConsent()` : Vérifie si un consentement valide existe

### Handlers

#### ConsentHandler
- `POST /consents` : Ajouter un consentement
- `GET /consents` : Récupérer tous les consentements de l'utilisateur
- `GET /consents/active?type=...` : Récupérer le consentement actif
- `GET /consents/check?type=...` : Vérifier si un consentement valide existe
- `POST /consents/{id}/revoke` : Révoquer un consentement

## Intégration avec la Détection d'Images Similaires

### Vérification automatique
La fonction `DetectSimilarMedia()` vérifie automatiquement le consentement avant de procéder :

```go
// Vérification du consentement RGPD
hasConsent, err := s.ConsentService.HasValidConsent(userID, "image_similarity_detection")
if err != nil {
    return nil, fmt.Errorf("erreur lors de la vérification du consentement RGPD : %v", err)
}
if !hasConsent {
    return nil, fmt.Errorf("consentement RGPD requis pour la détection d'images similaires")
}
```

### Suppression automatique des données
Lors de la suppression d'une image, le pHash associé est automatiquement supprimé de la base de données.

## Utilisation de l'API

### 1. Ajouter un consentement

```bash
POST /consents
Content-Type: application/json
Authorization: Bearer <token>

{
  "consent_type": "image_similarity_detection",
  "consent_version": "1.0",
  "is_granted": true,
  "consent_text": "Je consens au traitement algorithmique de mes images..."
}
```

### 2. Vérifier un consentement

```bash
GET /consents/check?type=image_similarity_detection
Authorization: Bearer <token>
```

### 3. Récupérer tous les consentements

```bash
GET /consents
Authorization: Bearer <token>
```

### 4. Révoquer un consentement

```bash
POST /consents/{id}/revoke
Authorization: Bearer <token>
```

## Conformité RGPD

### Principes respectés

1. **Consentement explicite** : L'utilisateur doit explicitement accepter le traitement
2. **Traçabilité** : Tous les consentements sont enregistrés avec horodatage et IP
3. **Révocation facile** : Possibilité de révoquer à tout moment
4. **Minimisation** : Seules les données nécessaires sont traitées
5. **Suppression automatique** : Les pHash sont supprimés avec les images
6. **Transparence** : Texte de consentement clair et complet

### Mesures de sécurité

- Enregistrement de l'IP et du User-Agent lors du consentement
- Horodatage de toutes les actions
- Vérification d'authentification sur toutes les routes
- Logs complets des opérations
- Suppression automatique des données associées

## Tests

### Scénarios de test recommandés

1. **Test d'ajout de consentement**
   - Ajouter un consentement valide
   - Vérifier qu'il est bien enregistré
   - Vérifier les métadonnées (IP, User-Agent, etc.)

2. **Test de vérification de consentement**
   - Vérifier qu'un consentement valide est détecté
   - Vérifier qu'un consentement révoqué n'est pas valide
   - Vérifier qu'un consentement inexistant retourne false

3. **Test de révocation**
   - Révoquer un consentement actif
   - Vérifier qu'il ne peut plus être utilisé
   - Vérifier les métadonnées de révocation

4. **Test d'intégration avec la détection d'images**
   - Tester sans consentement (doit échouer)
   - Tester avec consentement valide (doit réussir)
   - Tester après révocation (doit échouer)

## Migration de base de données

La migration ajoute automatiquement la table `consents` lors du démarrage de l'application :

```go
err := manager.DB.AutoMigrate(
    &models.User{},
    &models.Album{},
    &models.Media{},
    &models.Consent{}, // Nouvelle table
)
```

## Monitoring et Audit

### Logs générés

- Création de consentement
- Vérification de consentement
- Révocation de consentement
- Échecs de vérification
- Erreurs de base de données

### Métriques recommandées

- Nombre de consentements accordés par type
- Nombre de révocations
- Taux de consentement pour la détection d'images
- Temps de réponse des vérifications

## Évolutions futures

### Améliorations possibles

1. **Chiffrement des pHash** : Chiffrer les empreintes numériques en base
2. **Expiration automatique** : Ajouter une date d'expiration aux consentements
3. **Notifications** : Notifier l'utilisateur avant expiration
4. **Export RGPD** : Permettre l'export des données de consentement
5. **Dashboard admin** : Interface d'administration des consentements

### Extensions

- Support d'autres types de consentement
- Intégration avec un système de gestion des préférences
- API pour les demandes d'accès aux données (Article 15 RGPD)
- Système de notification des violations de données 