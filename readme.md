![Image](https://user-images.githubusercontent.com/35310806/131912718-c9d80cbb-e176-4a73-a2e3-03868c904b7c.png)


## Anwednung

Bei Fragen, bitte erstelle ein `Issue`, in dem du mich Tagst.

1. Meld dich bei Edubase Reader an und wähle das Buch/ Dokument aus, dass du umwandeln möchtest. 
2. Navigiere zu ersten Seite des Dokuments.
3. Öffne die Entwicklerkonsole des Browsers (Firefox / Google Chrome: F12) und Navigiere unter den Punkt `Console`.
4. Füge den gesamten Code der `pageDownloader.js` in die Konsole ein.
5. Setze für die Variable `maxPages` die Gesamtseitenzahl des Dokuments (diese wird im Unteren Bereich von Edubase angezeigt).
6. Das Script wird nun diverse HTML Dateien herunterladen und im Downloads Ordner speichern.
7. Nachdem alle HTML Dateien Heruntergeladen wurden, kopiere alle HTML Dateien in einen seperaten Ordner. 
8. Installiere [Node.js](https://nodejs.org/en/download/prebuilt-installer) auf Windows.
9. Installiere Yarn mit dem Windows CMD: 
``` cmd
npm install --global yarn
```
10. Erstelle einen Ordner, in der die Umgebung Installiert wird. Gehe mit ``` cmd
cd C:\$Pfad in den Ordner ```
11. Downloade die Dateien `package.json` und `yarn.lock` und bewege diese in den Ordner. Installiere die erforderlichen Datein mit ``` cmd
yarn install ```
12. Erstelle einen Ordner namens "Files" 
13. Kopiere alle Dateien (Page 1 - PageXY) in den Files Ordner.
14. Editiere die Datei `index.js` in einem Texteditor z.B Notepad und setze die Variable  `const numberOfPages = XY;` auf die Anzahl der Seiten, die gedownloaded wurden. Zudem editiere die Variable ```javascript
   Page.navigate({
     url: `file:///C:/Users/iamcool/Edubase_to_pdf/Files/page-${page}.html`,
   });  
   ```
   Der oben gezeigte Pfad zeigt nun auf den Ordner `C:/Users/iamcool/Edubase_to_pdf/Files`
15. Führe im CMD den Befehl `node .` aus.
16. Nun sollten im Ordner `pages` alle PDF Dateien gespeichert werden. Diese können nun mit einem PDF Programm zusammengeführt werden. z.B. Adobe Acrobat Reader.
17. Lass die Texterkennung über die PDF Datei laufen, um im PDF suchen und Kopieren zu können.

Falls es Probleme oder Fragen gibt, erstelle bitte ein `Issue` auf Github.


