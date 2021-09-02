"use strict";
const maxPages = 416;

const pages = [];
const btnNextPage = document.querySelector('[data-action="footer-next-page"]');
function triggerEvent(element, eventType) {
  if (element.fireEvent) {
    element.fireEvent("on" + eventType);
  } else {
    const event = document.createEvent("Events");
    event.initEvent(eventType, true, false);
    element.dispatchEvent(event);
  }
}
function downloadSVG(svg, pageNumber) {
  //get svg source.
  const serializer = new XMLSerializer();
  let source = serializer.serializeToString(svg);
  //add name spaces.
  if (!source.match(/^<svg[^>]+xmlns="http\:\/\/www\.w3\.org\/2000\/svg"/)) {
    source = source.replace(/^<svg/, '<svg xmlns="http://www.w3.org/2000/svg"');
  }
  if (!source.match(/^<svg[^>]+"http\:\/\/www\.w3\.org\/1999\/xlink"/)) {
    source = source.replace(
      /^<svg/,
      '<svg xmlns:xlink="http://www.w3.org/1999/xlink"'
    );
  }
  //add xml declaration
  source = '<?xml version="1.0" standalone="no"?>\r\n' + source;
  //convert svg source to URI data scheme.
  const url = "data:image/svg+xml;charset=utf-8," + encodeURIComponent(source);
  download(url, pageNumber, "svg");
}
function downloadHTML(htmlContent, pageNumber) {
  download(
    "data:text/html;charset=utf-8," + encodeURIComponent(htmlContent),
    pageNumber,
    "html"
  );
}
function download(url, pageNumber, extension) {
  const downloadLink = document.createElement("a");
  downloadLink.href = url;
  downloadLink.download = `page-${pageNumber}.${extension}`;
  document.body.appendChild(downloadLink);
  downloadLink.click();
  document.body.removeChild(downloadLink);
}
let previousPageContent;
let page = 1;
function processPage() {
  if (page > maxPages) {
    return;
  }
  setTimeout(() => {
    console.log(`Processing page ${page}/${maxPages}...`);
    const backgroundImage = document.querySelector(
      ".lu-page-background-image"
    ).src;
    let content = document.querySelector(".lu-page-svg-container svg");
    if (content && content !== null) {
      while (content !== null && content.innerHTML === previousPageContent) {
        console.log(`Content has not changed... try again in 1s...`);
        setTimeout(() => {
          content = document.querySelector(".lu-page-svg-container svg");
        }, 1000);
      }
      previousPageContent = content.innerHTML;
    } else {
      content = null;
    }
    downloadSVG(content, page);
    downloadHTML(
      `<html>
        <style>  
          img {
            background-image: url("${backgroundImage}");
          }
        </style>
        <img src="./page-${page}.svg" alt="" height="2721" width="1928" />
      </html>
  `,
      page
    );
    console.log(`Processing page ${page}/${maxPages} done. ✔`);
    triggerEvent(btnNextPage, "click");
    page = page + 1;
    processPage();
  }, 1500);
}

processPage();