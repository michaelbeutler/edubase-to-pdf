# Edubase Reader to PDF

Recently I had to buy some books for school and the books were even available as eBooks. Unfortunately, the eBooks are only accessible with the Edubase Reader. This motivated me to write an app that converts the Edubase SVG solution into a searchable PDF.

I wrote this script in the evening and it was not meant to be on GitHub but after I finished, I decided to upload it anyway. So, the instructions and the code are loose. Maybe I will rewrite the whole script at some point but not today.

## Usage

If you got some questions. Feel free to create a issue and linking me.

1. Login to your account and open the book you want to convert.
2. Navigate to the first page.
3. Open the development menu so you can access the console. (F12 on Windows)
4. Set the variable `maxPages` to the maximum amount pages your desired book has.
5. Paste the whole code into the console.
6. The script should start downloading lots of files so maybe you have to grant special permissions (Tested on Google Chrome).
7. Install the dependencies with `yarn`.
8. Set the variable `numberOfPages` inside `index.js` to the maximum amount pages your desired book has.
9. Set the property `url` inside `index.js` to the path your downloaded HTML files are avaliable and include the `page` variable.
   ```javascript
   Page.navigate({
     url: `file:///C:/Users/iamcool/Downloads/pages/page-${page}.html`,
   });
   ```
10. Run the script with `$ node .` and wait for the script to finish.
11. When everything is done, each page should be in a seperate file inside your `pages` directory.
12. Combinde the files with any programm you like. (I used Adobe Acrobat Reader.)
13. Let the app scan the PDF and detect texts. (I used Adobe Acrobat Reader.)
