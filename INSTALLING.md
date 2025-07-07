# Installation Guide
The project is split into 2 main parts; the `client/` and the `server/`. The client uses React (Typescript) + Tailwind (and utilises Vite as a build tool), while the `server/`, uses Golang. The server
contains: 
- (A) code that is relevant to the AI, which uses `Qdrant` (library; `GRPC` is used to communicate with it) and `HTTP` to communicate with Mistral and the embedding model.
- (B) code to help `client/` interact with the server (including the AI), which uses `Echo`.

In other words, you **DO NOT** need to install `client/` if you don't want to use the website. You can just use `server/` and then `go run cmd/cli.go` inside server/.

## Requirements
#### `server/`
- [Golang](https://go.dev/doc/install): We use version `1.21.4`.
- [Qdrant](https://github.com/qdrant/qdrant/releases/tag/v1.14.1): Our vector database. You must keep it running (e.g by opening another terminal/command prompt and typing qdrant inside).
- [Ollama](https://ollama.com/download/windows): LLM manager that allows us to download our other LLMs like Mistral easily.
- **Mistral:** Simply run `ollama pull mistral` after installing ollama. We use the 7B model; version 0.3.
- **Embedding model:** Simply run `ollama pull mxbai-embed-large` after installing ollama. We use the latest version.

#### `client/`
- [NodeJS](https://nodejs.org/en/download)
- After installing node and `cloning` this repo, simply navigate (via cd command) to the `client/` folder and run `npm install`. After it is completed, run `npm run dev` and copy paste
the URL provided in the terminal/command prompt inside your browser.

## Setting Up

**First**, make sure you complete all the requirements mentioned. After that, **make sure to clone the github repo using `git clone` first!** 
Either that, or you can Download as ZIP on github, then extract the source code in a folder.

Then, first navigate to `server/` (using the `cd` command) and run `go run cmd/setup.go`. When it asks you for input, first enter `0`, and after it is finished, use `go run cmd/setup.go`
again, and this time enter `1` (not the most intuitive design I know lol).

After that, if you intend to just talk to the AI directly in the terminal/command prompt, run `go run cmd/cli.go`. If you're also running your frontend in the `client/` folder (using
`npm run dev` like it is mentioned in the requirements), then run `go run cmd/server.go` for the frontend to be able to communicate with the backend.
