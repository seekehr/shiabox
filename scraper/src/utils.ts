import * as readline from "readline";

export function prompt(question: string): Promise<string> {
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
    return new Promise((resolve) => {
        rl.question(question, (answer) => { rl.close(); resolve(answer); });
    });
}

export function normalize(s: string): string {
    return s.normalize("NFD").replace(/[̀-ͯ]/g, "").toLowerCase().trim();
}

export function slugify(s: string): string {
    return normalize(s).replace(/\s+/g, "-").replace(/[^a-z0-9-]/g, "");
}

export async function fetchHtml(url: string, retries = 3): Promise<string | null> {
    for (let attempt = 1; attempt <= retries; attempt++) {
        try {
            const res = await fetch(url, {
                headers: { "User-Agent": "Mozilla/5.0 (compatible; shiabox-scraper/1.0)" },
            });
            if (res.ok) return res.text();
            if (res.status === 404) return null;
            if (attempt < retries) await new Promise((r) => setTimeout(r, attempt * 500));
        } catch {
            if (attempt < retries) await new Promise((r) => setTimeout(r, attempt * 500));
        }
    }
    console.error(`\nSkipping (failed after ${retries} attempts): ${url}`);
    return null;
}

/** Run tasks with at most `limit` running concurrently. */
export async function pool<T>(tasks: (() => Promise<T>)[], limit: number): Promise<T[]> {
    const results: T[] = new Array(tasks.length);
    let idx = 0;
    async function worker() {
        while (idx < tasks.length) {
            const i = idx++;
            results[i] = await tasks[i]();
        }
    }
    await Promise.all(Array.from({ length: Math.min(limit, tasks.length) }, worker));
    return results;
}
