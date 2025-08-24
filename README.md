# 🚀 AutoHost CLI

**Recupera el control de tus servicios.**  
**AutoHost CLI** es una herramienta de línea de comandos para instalar, configurar y administrar aplicaciones y servicios **en tu propio servidor**, sin depender de terceros y con un flujo de trabajo sencillo y automatizado.

---

## 🌟 Características

- **Instalación en un comando**: Despliega aplicaciones listas para usar con `app install`.
- **Configuración automática**: Ajusta dominios, certificados SSL y redes internas sin configuraciones manuales.
- **Soporte para múltiples apps**: Nextcloud, BookStack, y más (¡en constante crecimiento!).
- **Integración con Tailscale**: Conéctate de forma segura a tu infraestructura privada.
- **Compatibilidad con Docker**: Aislamiento y portabilidad de tus aplicaciones.
- **Enfoque en privacidad y control**: Todo se ejecuta en **tu** infraestructura.

---

## 📦 Instalación

Instala AutoHost CLI directamente desde GitHub con un solo comando:

```bash
curl -fsSL https://raw.githubusercontent.com/mazapanuwu13/autohost-cli/main/install.sh | bash
```

Este script detecta automáticamente tu sistema operativo y arquitectura, descarga la versión más reciente del binario desde GitHub Releases e instala AutoHost CLI en tu sistema.

---

## 🛠 Uso Básico

### Inicializar AutoHost
```bash
autohost init
```
### Configuracion inicial
```bash
autohost setup
```

### Instalar una aplicación
```bash
autohost app install bookstack
```

### Levantar una app
```bash
autohost app start bookstack
```

---

## 🔒 Filosofía

En un mundo donde la mayoría de las aplicaciones están en la nube, **AutoHost CLI** te devuelve el poder:  
- Controlas **tus datos**.  
- Eliminas la dependencia de múltiples SaaS.  
- Construyes tu propia infraestructura, escalable y privada.  


---

## 🤝 Contribuir

¿Quieres aportar?  
1. Haz un fork del repositorio.  
2. Crea una rama para tu feature/fix.  
3. Envía un Pull Request.  

---

## 📜 Licencia

Este proyecto está bajo la licencia **MIT**.

---

> 💡 **Consejo:** Si quieres recibir actualizaciones y novedades, visita [authost.dev](https://autohst.dev) o síguenos en redes.