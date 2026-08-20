---
sidebar_position: 7
title: Weboberfläche
---

# Weboberfläche

Verwalte alle Konfigurationen visuell im Browser. Der Daemon startet bei Bedarf automatisch.

## Verwendung

```bash
# Im Browser öffnen (startet den Daemon bei Bedarf)
zen web
```

## Funktionen

- Verwaltung von Anbietern und Profilen
- Verwaltung der Projektbindungen
- Globale Einstellungen (Standardclient, Standardprofil, Ports)
- Einstellungen für die Konfigurationssynchronisierung
- Anzeige der Anfrageprotokolle mit automatischer Aktualisierung
- Autovervollständigung im Modellfeld

## Sicherheit

Beim ersten Start des Daemons wird automatisch ein Zugangspasswort erzeugt. Nicht-lokale Anfragen (außerhalb von 127.0.0.1/::1) erfordern eine Anmeldung.

- Sitzungsbasierte Authentifizierung mit konfigurierbarer Ablaufzeit
- Schutz vor Brute-Force-Angriffen mit exponentiell wachsender Wartezeit
- RSA-Verschlüsselung für den Transport sensibler Tokens (API-Schlüssel werden im Browser verschlüsselt)
- Lokale Zugriffe (127.0.0.1) umgehen die Authentifizierung

### Passwortverwaltung

```bash
# Passwort der Weboberfläche zurücksetzen
zen config reset-password

# Passwort in der Weboberfläche ändern
zen web  # Einstellungen → Passwort ändern
```
