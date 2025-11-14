# Lojagtec

This is a project is an online shop for a regional business that sells water filters, water fountains, and services related to their maintenance.

## Tech Stack

*   **Backend:** Go
*   **Frontend:** HTMX, Tailwind CSS
*   **Database:** PostgreSQL

## Getting Started

### Prerequisites

*   [Go](https://golang.org/doc/install)
*   [PostgreSQL](https://www.postgresql.org/download/)

### Installation

1.  **Clone the repository:**

    ```bash
    git clone <repository-url>
    cd lojagtec
    ```

2.  **Build the CSS:**

    ```bash
    tailwind -i ./web/static/css/style.css -o ./web/static/css/dist/style.css -w
    ```

    This command will watch for changes in `web/static/css/style.css` and rebuild the `web/static/css/dist/style.css` file.

4.  **Set up the database:**

    *   Create a PostgreSQL database.
    *   Update the database connection string in `configs/config.toml` (you will need to create this file).

5.  **Run the application:**

    ```bash
    go run cmd/server/main.go
    ```

    The application will be available at `http://localhost:8080`.

## Project Structure

```
.
├── cmd
│   └── server
│       └── main.go
├── configs
├── go.mod
├── internal
│   ├── database
│   ├── handlers
│   └── models
├── package.json
├── postcss.config.js
├── README.md
├── scripts
├── tailwind.config.js
└── web
    ├── static
    │   ├── css
    │   │   ├── dist
    │   │   └── style.css
    │   ├── images
    │   └── js
    └── templates
        └── index.html
```
