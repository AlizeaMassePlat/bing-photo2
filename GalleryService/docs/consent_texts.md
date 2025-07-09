# Textes de Consentement RGPD

## Consentement pour la Détection d'Images Similaires

### Version 1.0 - Texte de Consentement

**Titre : Consentement au traitement algorithmique de vos images pour la détection de doublons**

**Description du traitement :**

Nous vous demandons votre consentement pour analyser vos images afin de détecter automatiquement les doublons et les images similaires dans votre galerie. Cette fonctionnalité vous permet d'optimiser votre espace de stockage et d'organiser plus efficacement vos photos.

**Données traitées :**
- Vos images personnelles stockées dans votre galerie
- Empreintes numériques (pHash) calculées à partir du contenu de vos images
- Métadonnées techniques de vos images (taille, format, nom)

**Finalité du traitement :**
- Détection automatique de doublons exacts
- Identification d'images visuellement similaires
- Optimisation de l'espace de stockage
- Amélioration de l'expérience utilisateur

**Base légale :**
Votre consentement explicite (Article 6.1.a du RGPD)

**Durée de conservation :**
- Les empreintes numériques sont conservées tant que l'image correspondante existe dans votre galerie
- Suppression automatique lors de la suppression de l'image
- Possibilité de révoquer votre consentement à tout moment

**Destinataires :**
- Nos serveurs sécurisés uniquement
- Aucun partage avec des tiers
- Traitement effectué localement sur nos infrastructures

**Vos droits :**
- Droit de retirer votre consentement à tout moment
- Droit d'accéder à vos données
- Droit de supprimer vos données
- Droit de porter vos données
- Droit de déposer une plainte auprès de la CNIL

**Sécurité :**
- Chiffrement des empreintes numériques
- Accès restreint aux données
- Audit trail complet des traitements
- Conformité aux standards de sécurité

**Contact :**
Pour toute question concernant ce traitement, contactez notre délégué à la protection des données à l'adresse : dpo@votreentreprise.com

---

### Version 1.1 - Texte de Consentement (Mise à jour)

**Titre : Consentement au traitement algorithmique de vos images pour la détection de doublons et l'organisation intelligente**

**Description du traitement :**

Nous vous demandons votre consentement pour analyser vos images afin de détecter automatiquement les doublons et les images similaires dans votre galerie. Cette fonctionnalité vous permet d'optimiser votre espace de stockage, d'organiser plus efficacement vos photos et de découvrir des groupes d'images similaires.

**Données traitées :**
- Vos images personnelles stockées dans votre galerie
- Empreintes numériques (pHash) calculées à partir du contenu de vos images
- Métadonnées techniques de vos images (taille, format, nom, date de création)
- Informations de similarité entre vos images

**Finalité du traitement :**
- Détection automatique de doublons exacts
- Identification d'images visuellement similaires
- Regroupement automatique d'images similaires
- Optimisation de l'espace de stockage
- Amélioration de l'expérience utilisateur et de l'organisation

**Base légale :**
Votre consentement explicite (Article 6.1.a du RGPD)

**Durée de conservation :**
- Les empreintes numériques sont conservées tant que l'image correspondante existe dans votre galerie
- Suppression automatique lors de la suppression de l'image
- Possibilité de révoquer votre consentement à tout moment
- Conservation des logs de traitement pendant 12 mois maximum

**Destinataires :**
- Nos serveurs sécurisés uniquement
- Aucun partage avec des tiers
- Traitement effectué localement sur nos infrastructures
- Pas d'utilisation pour d'autres finalités

**Vos droits :**
- Droit de retirer votre consentement à tout moment
- Droit d'accéder à vos données
- Droit de supprimer vos données
- Droit de porter vos données
- Droit de déposer une plainte auprès de la CNIL
- Droit d'opposition au traitement

**Sécurité :**
- Chiffrement des empreintes numériques en base de données
- Accès restreint aux données avec authentification
- Audit trail complet des traitements effectués
- Conformité aux standards de sécurité ISO 27001
- Sauvegarde sécurisée des données

**Contact :**
Pour toute question concernant ce traitement, contactez notre délégué à la protection des données à l'adresse : dpo@votreentreprise.com

---

## Utilisation des Textes

### Pour l'API

Lors de l'ajout d'un consentement via l'API, utilisez le texte correspondant à la version souhaitée :

```json
{
  "consent_type": "image_similarity_detection",
  "consent_version": "1.0",
  "is_granted": true,
  "consent_text": "[Texte complet du consentement version 1.0]"
}
```

### Types de Consentement Supportés

- `image_similarity_detection` : Détection d'images similaires
- `data_processing` : Traitement général des données personnelles
- `analytics` : Analyses et statistiques d'usage
- `marketing` : Communications marketing (si applicable)

### Gestion des Versions

- Chaque mise à jour du texte de consentement doit avoir un nouveau numéro de version
- Les anciens consentements restent valides jusqu'à révocation
- Les nouveaux utilisateurs reçoivent automatiquement la dernière version
- Possibilité de demander le consentement à nouveau pour les mises à jour importantes 