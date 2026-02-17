---
sidebar_position: 7
title: Interfaz Web
---

# Interfaz Web

Gestiona visualmente todas las configuraciones a través de tu navegador. El daemon se inicia automáticamente cuando es necesario.

## Uso

```bash
# Open in browser (auto-starts daemon if needed)
zen web
```

## Características

- Gestión de proveedores y perfiles
- Gestión de vinculaciones de proyecto
- Configuración global (cliente predeterminado, perfil predeterminado, puertos)
- Ajustes de sincronización de configuración
- Visor de registros de solicitudes con actualización automática
- Autocompletado del campo de modelo

## Seguridad

Al iniciar el daemon por primera vez, se genera automáticamente una contraseña de acceso. Las solicitudes no locales (fuera de 127.0.0.1/::1) requieren inicio de sesión.

- Autenticación basada en sesiones con expiración configurable
- Protección contra fuerza bruta con retroceso exponencial
- Cifrado RSA para transporte de tokens sensibles (claves API cifradas en el navegador)
- El acceso local (127.0.0.1) no requiere autenticación

### Password Management

```bash
# Reset the Web UI password
zen config reset-password

# Change password via Web UI
zen web  # Settings → Change Password
```
