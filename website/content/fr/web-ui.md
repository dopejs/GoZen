---
sidebar_position: 7
title: Interface web
---

# Interface web

Gérez visuellement toutes les configurations depuis votre navigateur. Le daemon démarre automatiquement quand c’est nécessaire.

## Utilisation

```bash
# Ouvrir dans le navigateur (démarre le daemon si besoin)
zen web
```

## Fonctionnalités

- Gestion des fournisseurs et des profils
- Gestion des liaisons de projet
- Réglages globaux (client par défaut, profil par défaut, ports)
- Réglages de synchronisation de la configuration
- Visualiseur de journaux de requêtes avec rafraîchissement automatique
- Autocomplétion du champ modèle

## Sécurité

Au premier démarrage du daemon, un mot de passe d’accès est généré automatiquement. Les requêtes non locales (hors 127.0.0.1/::1) exigent une connexion.

- Authentification par session, avec expiration configurable
- Protection contre la force brute, avec temporisation exponentielle
- Chiffrement RSA pour le transport des jetons sensibles (clés d’API chiffrées dans le navigateur)
- L’accès local (127.0.0.1) contourne l’authentification

### Gestion du mot de passe

```bash
# Réinitialiser le mot de passe de l’interface web
zen config reset-password

# Le changer depuis l’interface web
zen web  # Réglages → Changer le mot de passe
```
