# 🌀 SwagFluence

***Disclaimer: This README was partially/fully generated with the assistance of AI.***

SwagFluence is a **Golang-based CLI tool** that automatically converts a Swagger/OpenAPI specification into beautifully formatted **Confluence documentation**. It generates:

* A parent documentation hub
* Individual pages for each API endpoint
* Detailed request/response info
* Auto-generated example JSON
* Clean tables, layout macros, and status tags

SwagFluence supports **Swagger 2.0** and **OpenAPI 3.x**, including `$ref` schema resolution.

---

## ✨ Features

### ✔️ Full Swagger/OpenAPI Parsing

* Reads Swagger/OpenAPI JSON from any URL
* Extracts operations, parameters, request bodies, schemas, tags
* Supports both:

  * `components/schemas` (OpenAPI 3.x)
  * `definitions` (Swagger 2.0)

### ✔️ Automatic Confluence Page Generation

SwagFluence creates or updates:

* A **parent documentation page** for the API
* A **separate page for each endpoint**

Each endpoint page includes:

* Method badges (GET/POST/etc.)
* Description, tags, operation ID
* Parameter tables
* Request body breakdown
* Schema tables with constraints
* Auto-generated **Example JSON**
* Confluence storage-format markup
* Layout macros for clean presentation

### ✔️ Intelligent Naming

Endpoints use automatic title generation based on:

1. Operation summary
2. Operation ID
3. Humanized path segments

### ✔️ Local Preview Mode

If Confluence credentials are not set:
SwagFluence simply prints all generated documentation to the terminal.

---

## 📦 Requirements

* **Go 1.18+**
* A reachable Swagger/OpenAPI endpoint
* Optional: Confluence REST API credentials

---

## 🔧 Installation

```bash
git clone <your-repo-url>
cd <repo>
go mod tidy
go build -o SwagFluence main.go

# Or cross-compile from another OS
GOOS=linux GOARCH=amd64 go build -o SwagFluence main.go
```
---

## 🚀 Usage

### **Default Mode (No Confluence Upload)**

```bash
./SwagFluence https://petstore.swagger.io/v2/swagger.json
```

This will:

* Fetch Swagger/OpenAPI JSON from:
  `https://petstore.swagger.io/v2/swagger.json`
* Generate full Confluence-ready documentation
* Print it to the terminal

---

## 🧩 Confluence Integration

Set the following environment variables:

```bash
export CONFLUENCE_BASE_URL="https://yourcompany.atlassian.net/wiki"
export CONFLUENCE_USERNAME="you@company.com"
export CONFLUENCE_API_TOKEN="API_TOKEN"
export CONFLUENCE_SPACE_KEY="ENG"
export CONFLUENCE_PARENT_PAGE_ID="123456"   # optional
```

Run:

```bash
./SwagFluence https://petstore.swagger.io/v2/swagger.json
```

SwagFluence will:

1. Create/update the parent page
2. Create/update one page per endpoint
3. Output links to all generated pages

---

## 🏗 Project Structure

```
.
├── main.go
├── go.mod
└── README.md
```

---

## 🧠 How SwagFluence Works

1. **Fetch Swagger/OpenAPI JSON**
2. **Parse paths, operations, and schemas**
3. **Generate Confluence-compatible markup**, including:

   * Tables
   * Layout sections
   * Status macros
   * Auto-built Example JSON
4. **Create or update** Confluence pages using REST API

---

## 🤝 Contributing

Contributions welcome!
Ideas for enhancements:

* Response schema documentation
* YAML Swagger support

---

## 📄 License

This software is licensed under the [Apache License 2.0](./LICENSE).
