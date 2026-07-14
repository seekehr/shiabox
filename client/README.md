# shiabox client

Vite + React + TypeScript frontend for searching ahadith via the FastAPI backend.

```bash
npm install
npm run dev
```

Point `BACKEND_URL` in [`src/controller/controllers.ts`](src/controller/controllers.ts) at your API (default `http://localhost:1323`). Hadith data is scraped and indexed by [`scraper/`](../scraper/) and [`server/`](../server/); the client only talks to the API.
