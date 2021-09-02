const chromeLauncher = require("chrome-launcher");
const CDP = require("chrome-remote-interface");
const fs = require("fs");

const numberOfPages = 416;

const launchChrome = () =>
  chromeLauncher.launch({
    chromeFlags: ["--disable-gpu", "--headless"],
  });

async function main() {
  fs.mkdirSync("pages");
  for (let page = 1; page <= numberOfPages; page++) {
    const chrome = await launchChrome();
    const protocol = await CDP({ port: chrome.port });
    const { Page } = protocol;

    await Page.enable();
    Page.navigate({
      url: `ADD_YOUR_PATH`,
    });

    Page.loadEventFired(async () => {
      const { data } = await Page.printToPDF({
        scale: 0.5,
        landscape: false,
        printBackground: true,
        displayHeaderFooter: false,
        transferMode: "ReturnAsBase64",
        paperWidth: 8.3,
        paperHeight: 11.7,
      });
      fs.writeFileSync(`./pages/page${page}.pdf`, Buffer.from(data, "base64"));

      chrome.kill();
    });
  }
}

main();
