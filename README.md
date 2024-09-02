# Edubase-to-PDF CLI Tool 🎓📚

## Description 📚🖨📑
The `edubase-to-pdf` CLI tool is designed to help users securely download and archive e-books from Edubase as PDF files. 📖🔒 It allows users to access their books even if the Edubase reader gets discontinued, ensuring continued access to educational resources. Please note that this tool is intended strictly for personal use and archiving purposes. It should not be used for any illegal activities, including piracy. 🚫🏴‍☠️


## 🎬 Demo

Check out this demo to see edubase-to-pdf in action! 👇

![Demo](demo.gif)

## 🌟 Features

- 🔍 **Easy**: Use one single tool to download all your eBooks.
- 📚 **PDF**: Save your eBooks as PDF files for easy access.
- 📧 **Secure**: Log in securely using your Edubase email and password.
- ➡ **Customizable**: Choose the starting page and the number of pages to import.
- 📂 **Temporary Directory**: Specify a temporary directory for screenshots.
- ⏳ **Page Delay**: Set a delay between pages to give the browser time to load.
- 🔎 **Browser Size**: Customize the browser width and height for better screenshot quality.

## 📦 Installation

### 🔧 Binaries

You can install the edubase-to-pdf binary easily using the following command:

```zsh
# This will install the binary at $(go env GOPATH)/bin/edubase-to-pdf
curl -sSfL https://raw.githubusercontent.com/michaelbeutler/edubase-to-pdf/main/install.sh | sh -s -- -b $(go env GOPATH)/bin

# ✅ Verify the installation by checking the help
edubase-to-pdf --help
```

### 🖥️ Windows

For Windows users, you can install the edubase-to-pdf binary using Chocolatey:

```powershell
# Install using Chocolatey
choco install michaelbeutler-edubase-to-pdf --version=2.0.3

# ✅ Verify the installation by checking the help
edubase-to-pdf --help
```

### 🐳 Docker

You can also run the edubase-to-pdf using Docker:

```sh
# Pull the latest Docker image
docker pull ghcr.io/michaelbeutler/edubase-to-pdf

# Run the Docker container
docker run -it ghcr.io/michaelbeutler/edubase-to-pdf edubase-to-pdf --help

# Run the Docker container to import a book
docker run -v ./ ghcr.io/michaelbeutler/edubase-to-pdf edubase-to-pdf import
```

## Example 🧾👆

Here is an example of how to use the tool:

```shell
edubase-to-pdf import -e your_email@example.com -p your_password -s 2 -m 10
```

In this example, the tool signs in to Edubase using the provided email and password. It then starts importing from page 2 and imports a maximum of 10 pages. The resulting PDF will be saved in the current directory. 🎉📚

## Contact 🤔💬

If you encounter any issues or have any questions, please feel free to open an issue on our GitHub repository:

[github.com/michaelbeutler/edubase-to-pdf/issues](https://github.com/michaelbeutler/edubase-to-pdf/issues)

We value your feedback and will do our best to assist you. 👍📧

## Usage 💻⌨

```shell
edubase-to-pdf import [flags]
```

## Flags 🚩

```shell
  -d, --debug                 Debug mode. Show browser window.
  -e, --email string          Edubase email for login. 📧
  -H, --height int            Browser height in pixels; this can affect screenshot quality. (default 1440) 🔍
  -h, --help                  Help for import.
  -m, --max-pages int         Maximum pages to import from the book. (default -1) 🔝
  -D, --page-delay duration   Delay between pages in milliseconds. This is required to give the browser time to load the page. (default 500ms) ⏳
  -p, --password string       Edubase password for login. 🔑
  -s, --start-page int        Start page to import from the book. (default 1) ➡
  -t, --temp string           Temporary directory for screenshots; these will be used to generate the pdf. (default "screenshots") 📂
  -W, --width int             Browser width in pixels; this can affect screenshot quality. (default 2560) 🔎
  -T, --timeout duration      Maximum time the app can take to download all pages. (increase this value for large books, default 5 min)
```

## Legal Disclaimer ⚖️

**Please note that the `edubase-to-pdf` CLI tool is not affiliated with Edubase and should be used responsibly and within the bounds of the law.** This tool is intended solely for personal use, archiving purposes, and accessing books in compliance with the terms and conditions set by Edubase. The tool should not be used to infringe upon the copyrights or intellectual property rights of any individual or organization. The developer of this tool disclaims any liability for any misuse or illegal activities performed with it. Users are solely responsible for their actions while using this tool. 🚫👮‍♂️

Remember to respect the rights of authors and publishers by using this tool responsibly and legally. Happy reading! 📚😊