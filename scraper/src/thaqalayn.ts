import { chromium } from "playwright";
import * as readline from "readline";

function prompt(question: string): Promise<string> {
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
    return new Promise((resolve) => {
        rl.question(question, (answer) => {
            rl.close();
            resolve(answer);
        });
    });
}

function normalize(s: string): string {
    return s
        .normalize("NFD")           // decompose diacritics (ā → a + combining macron)
        .replace(/[̀-ͯ]/g, "") // strip combining marks
        .toLowerCase()
        .trim();
}

const bookName = normalize(await prompt("Enter book name: "));

const browser = await chromium.launch({ headless: false });
const page = await browser.newPage();

await page.goto("https://thaqalayn.net");

// Find the book card whose title matches (case-insensitive, trimmed)
const bookLinks = page.locator("a[href^='/book/']");
const count = await bookLinks.count();

let clicked = false;
for (let i = 0; i < count; i++) {
    const link = bookLinks.nth(i);
    const titleEl = link.locator("div.font-bold");
    const titleText = normalize((await titleEl.textContent()) ?? "");
    if (titleText === bookName || titleText.includes(bookName)) {
        console.log(`Clicking: "${await titleEl.textContent()}"`);
        await link.click();
        clicked = true;
        break;
    }
}

if (!clicked) {
    console.log(`No book found matching "${bookName}".`);
    await browser.close();
    process.exit(1);
}

await page.waitForLoadState("networkidle");
console.log(`Navigated to: ${page.url()}`);
