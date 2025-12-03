# Edubase-to-PDF CLI Tool 🎓📚

## Beschreibung 📚🖨📑
Das `edubase-to-pdf` CLI-Tool wurde entwickelt, um Nutzer:innen beim sicheren Herunterladen und Archivieren von E-Books aus Edubase als PDF-Dateien zu unterstützen. 📖🔒 Damit können Bücher auch dann noch genutzt werden, falls der Edubase-Reader eingestellt wird – so bleibt der Zugang zu Bildungsressourcen erhalten. Bitte beachte, dass dieses Tool ausschließlich für den **persönlichen Gebrauch und Archivierungszwecke** gedacht ist. Es darf nicht für illegale Aktivitäten, einschließlich Piraterie, genutzt werden. 🚫🏴‍☠️

## 🎬 Demo

Schau dir diese Demo an, um edubase-to-pdf in Aktion zu sehen! 👇

![Demo](demo.gif)

## 🌟 Funktionen

- 🔍 **Einfach**: Nutze ein einziges Tool, um alle deine eBooks herunterzuladen.  
- 📚 **PDF**: Speichere deine eBooks als PDF-Dateien für leichten Zugriff.  
- 📧 **Sicher**: Melde dich mit deiner Edubase-E-Mail und deinem Passwort sicher an.  
- ➡ **Anpassbar**: Wähle die Startseite und die Anzahl der zu importierenden Seiten.  
- 📂 **Temporäres Verzeichnis**: Gib ein temporäres Verzeichnis für Screenshots an.  
- ⏳ **Seiten-Verzögerung**: Lege eine Wartezeit zwischen den Seiten fest, damit der Browser laden kann.  
- 🔎 **Browsergröße**: Passe Breite und Höhe des Browsers an, um die Screenshot-Qualität zu verbessern.  
- 😵‍💫 **Leichtgewichtig**: Einzelne ausführbare Datei, kein Ballast wie Python-Skripte. 😉  

## 📦 Installation

### 🖼️ Installationsvideo
Für eine leichtere Installation und Nutzung gibt es ein Video:  

[YouTube Edubase-to-PDF Installations-Tutorial](https://youtu.be/BLNL_C_Bdbw)

### 🔧 Binaries

Installiere die `edubase-to-pdf`-Binary einfach mit folgendem Befehl:  

```zsh
# Dies installiert die Binary unter $(go env GOPATH)/bin/edubase-to-pdf
curl -sSfL https://raw.githubusercontent.com/michaelbeutler/edubase-to-pdf/main/install.sh | sh -s -- -b $(go env GOPATH)/bin

# ✅ Überprüfe die Installation mit:
edubase-to-pdf --help
```

### 🖥️ Windows

Für Windows-Nutzer:innen kann die Binary über Chocolatey installiert werden:  

```powershell
# Installation mit Chocolatey
choco install michaelbeutler-edubase-to-pdf --version=2.0.3

# ✅ Überprüfe die Installation mit:
edubase-to-pdf --help
```

### 🐳 Docker

Alternativ lässt sich `edubase-to-pdf` auch mit Docker ausführen:  

```sh
# Neuestes Docker-Image ziehen
docker pull ghcr.io/michaelbeutler/edubase-to-pdf

# Container starten
docker run -it ghcr.io/michaelbeutler/edubase-to-pdf edubase-to-pdf --help

# Container starten, um ein Buch zu importieren
docker run -v ./ ghcr.io/michaelbeutler/edubase-to-pdf edubase-to-pdf import
```

## Beispiel 🧾👆

So kannst du das Tool verwenden:  

```shell
edubase-to-pdf import -e deine_email@example.com -p dein_passwort -s 2 -m 10
```

In diesem Beispiel meldet sich das Tool mit der angegebenen E-Mail und dem Passwort bei Edubase an. Es beginnt ab Seite 2 und importiert maximal 10 Seiten. Das Ergebnis wird als PDF im aktuellen Verzeichnis gespeichert. 🎉📚

## Kontakt 🤔💬

Wenn du auf Probleme stößt oder Fragen hast, eröffne gerne ein Issue im GitHub-Repository:  

[github.com/michaelbeutler/edubase-to-pdf/issues](https://github.com/michaelbeutler/edubase-to-pdf/issues)

Dein Feedback ist willkommen – wir helfen dir so gut wie möglich. 👍📧

## Verwendung 💻⌨

```shell
edubase-to-pdf import [flags]
```

## Flags 🚩

```shell
  -d, --debug                 Debug-Modus. Browserfenster anzeigen.
  -M, --manual                Zugangsdaten manuell eingeben. Nützlich, wenn du Microsoft-Login nutzt oder den Entwickler:innen nicht vertraust 🪟.
  -e, --email string          Edubase-E-Mail für den Login. 📧
  -H, --height int            Browserhöhe in Pixeln; kann die Screenshot-Qualität beeinflussen. (Standard 1080) 🔍
  -h, --help                  Hilfe für import.
  -m, --max-pages int         Maximale Seitenzahl, die aus dem Buch importiert werden soll. (Standard -1) 🔝
  -o  --img-overwrite         Vorhandene Screenshots überschreiben. 🖼️
  -D, --page-delay duration   Verzögerung zwischen den Seiten in Millisekunden. Nötig, damit der Browser laden kann. (Standard 500ms) ⏳
  -p, --password string       Edubase-Passwort für den Login. 🔑
  -s, --start-page int        Startseite für den Import. (Standard 1) ➡
  -t, --temp string           Temporäres Verzeichnis für Screenshots, die zur PDF-Erstellung verwendet werden. (Standard "screenshots") 📂
  -W, --width int             Browserbreite in Pixeln; kann die Screenshot-Qualität beeinflussen. (Standard 1920) 🔎
  -T, --timeout duration      Maximale Zeit, die die App zum Download aller Seiten benötigt. (Für große Bücher erhöhen; Standard 5 Min.)
```

### ⚠️ Wichtiger Hinweis zur Bildschirmauflösung

**Mindestauflösung:** Das Tool benötigt eine Mindestauflösung von **1920x1080 (1080p)**, um ordnungsgemäß zu funktionieren. Bei niedrigeren Auflösungen kann es zu Problemen beim Erkennen der maximalen Seitenzahl kommen, da die Webseite zu stark vergrößert dargestellt wird.

- ✅ **Empfohlen:** 1920x1080 oder höher (Standard)
- ⚠️ **Nicht empfohlen:** Auflösungen unter 1920x1080

Falls du eine niedrigere Auflösung verwenden musst, kann das Tool eine Warnung ausgeben. In diesem Fall versuche, die Browser-Auflösung anzupassen:

```shell
edubase-to-pdf import -W 1920 -H 1080 -e deine_email@example.com -p dein_passwort
```

## Alternativen 🔄📚

- https://github.com/rtfmkiesel/edubase-downloader  
- gerne Pull Request eröffnen, um weitere Alternativen hinzuzufügen  

## Rechtlicher Hinweis ⚖️

**Bitte beachte: Das `edubase-to-pdf` CLI-Tool steht in keiner Verbindung zu Edubase und muss verantwortungsvoll und im Rahmen der gesetzlichen Bestimmungen genutzt werden.**  
Es dient ausschließlich dem persönlichen Gebrauch, zur Archivierung und zum Zugriff auf Bücher im Einklang mit den Nutzungsbedingungen von Edubase.  

Das Tool darf nicht zur Verletzung von Urheber- oder geistigen Eigentumsrechten verwendet werden. Der Entwickler übernimmt keinerlei Haftung für Missbrauch oder illegale Aktivitäten. Die Verantwortung liegt allein bei den Nutzer:innen. 🚫👮‍♂️  

Denke daran, die Rechte von Autor:innen und Verlagen zu respektieren – nutze das Tool verantwortungsbewusst und legal. Viel Spaß beim Lesen! 📚😊  
