import { chromium, Page } from "playwright";
import * as readline from "readline";
import * as fs from "fs";
import * as path from "path";

// ── Types ──────────────────────────────────────────────────────────────────

interface Hadith {
    book_name: string;
    book_number: number;
    chapter_number: number;
    chapter_name: string;
    hadith_number: number;
    arabic: string | null;
    english: string | null;
    url: string;
}

// ── Helpers ────────────────────────────────────────────────────────────────

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
        .normalize("NFD")
        .replace(/[̀-ͯ]/g, "")
        .toLowerCase()
        .trim();
}

function slugify(s: string): string {
    return normalize(s).replace(/\s+/g, "-").replace(/[^a-z0-9-]/g, "");
}

async function getText(page: Page, selector: string): Promise<string | null> {
    try {
        const el = page.locator(selector).first();
        if (await el.count() === 0) return null;
        return (await el.textContent())?.trim() ?? null;
    } catch {
        return null;
    }
}

// ── Main ───────────────────────────────────────────────────────────────────

const bookQuery = normalize(await prompt("Enter book name: "));

const browser = await chromium.launch({ headless: false });
const page = await browser.newPage();

// 1. Navigate to homepage and find the book
await page.goto("https://thaqalayn.net");
await page.waitForLoadState("networkidle");

const bookLinks = page.locator("a[href^='/book/']");
const count = await bookLinks.count();

let bookHref: string | null = null;
let bookNameActual = "";

for (let i = 0; i < count; i++) {
    const link = bookLinks.nth(i);
    const titleEl = link.locator("div.font-bold");
    if (await titleEl.count() === 0) continue;
    const raw = (await titleEl.textContent()) ?? "";
    if (normalize(raw).includes(bookQuery)) {
        bookNameActual = raw.trim();
        bookHref = await link.getAttribute("href");
        console.log(`Found: "${bookNameActual}" → ${bookHref}`);
        break;
    }
}

if (!bookHref) {
    console.log(`No book found matching "${bookQuery}".`);
    await browser.close();
    process.exit(1);
}

// 2. Create output file immediately
const outputDir = path.join(import.meta.dirname, "..", "output");
fs.mkdirSync(outputDir, { recursive: true });
const outputPath = path.join(outputDir, `${slugify(bookNameActual)}.json`);
fs.writeFileSync(outputPath, "[]", "utf-8");
console.log(`Output file created: ${outputPath}`);

// 3. Navigate to the book page and collect all chapter links
await page.goto(`https://thaqalayn.net${bookHref}`);
await page.waitForLoadState("networkidle");

const chapterLinks = await page.locator("a[href^='/chapter/']").all();
const chapterHrefs: string[] = [];
for (const link of chapterLinks) {
    const href = await link.getAttribute("href");
    if (href) chapterHrefs.push(href);
}
console.log(`Found ${chapterHrefs.length} chapter(s).`);

// 4. Parse book number from the first chapter href: /chapter/{bookId}/{bookNum}/{chapterNum}
const bookNumberMatch = chapterHrefs[0]?.match(/\/chapter\/\d+\/(\d+)\//);
const bookNumber = bookNumberMatch ? parseInt(bookNumberMatch[1]) : 0;

// 5. Scrape each chapter
const allHadiths: Hadith[] = [];

for (const chapterHref of chapterHrefs) {
    await page.goto(`https://thaqalayn.net${chapterHref}`);
    await page.waitForLoadState("networkidle");

    // Parse chapter metadata from the page
    const chapterMatch = chapterHref.match(/\/chapter\/\d+\/\d+\/(\d+)/);
    const chapterNumber = chapterMatch ? parseInt(chapterMatch[1]) : 0;

    // Chapter name: the heading between "Book X, Chapter Y" and the hadith list
    const chapterName = await getText(page, "h1, h2, h3") ?? "";

    // Collect all individual hadith links on this chapter page
    const hadithLinks = await page.locator("a[href^='/hadith/']").all();
    const hadithHrefs: string[] = [];
    for (const link of hadithLinks) {
        const href = await link.getAttribute("href");
        if (href && !hadithHrefs.includes(href)) hadithHrefs.push(href);
    }

    // If the chapter page already shows all ahadith inline (no separate hadith pages),
    // scrape them directly from the chapter page.
    if (hadithHrefs.length === 0) {
        // Scrape inline ahadith blocks
        const hadithBlocks = await page.locator("div[id^='hadith-'], section[id^='hadith-']").all();

        // Fallback: look for numbered hadith headings and associated paragraphs
        const headings = await page.locator("h2, h3, h4").all();
        let hadithNum = 0;
        for (const heading of headings) {
            const text = (await heading.textContent())?.trim() ?? "";
            if (!/^[Ḥh]ad[iī]th\s*#?\d+/i.test(text) && !/^#\d+/.test(text)) continue;
            hadithNum++;

            // Arabic: next sibling with Arabic text (RTL paragraph)
            const arabic = await heading.evaluate((el) => {
                let next = el.nextElementSibling;
                while (next) {
                    const dir = next.getAttribute("dir") ?? getComputedStyle(next).direction;
                    const text = next.textContent?.trim() ?? "";
                    if ((dir === "rtl" || /[؀-ۿ]/.test(text)) && text.length > 0)
                        return text;
                    next = next.nextElementSibling;
                }
                return null;
            });

            // English: next paragraph after Arabic
            const english = await heading.evaluate((el) => {
                let next = el.nextElementSibling;
                let seenArabic = false;
                while (next) {
                    const dir = next.getAttribute("dir") ?? getComputedStyle(next).direction;
                    const text = next.textContent?.trim() ?? "";
                    if ((dir === "rtl" || /[؀-ۿ]/.test(text)) && text.length > 0) {
                        seenArabic = true;
                    } else if (seenArabic && text.length > 0) {
                        return text;
                    }
                    next = next.nextElementSibling;
                }
                return null;
            });

            allHadiths.push({
                book_name: bookNameActual,
                book_number: bookNumber,
                chapter_number: chapterNumber,
                chapter_name: chapterName,
                hadith_number: hadithNum,
                arabic: arabic ?? null,
                english: english ?? null,
                url: `https://thaqalayn.net${chapterHref}`,
            });
        }

        console.log(`  Chapter ${chapterNumber}: scraped ${hadithNum} inline hadith(s).`);
        continue;
    }

    // 6. Visit each hadith page
    console.log(`  Chapter ${chapterNumber} "${chapterName}": ${hadithHrefs.length} hadith page(s).`);

    for (const hadithHref of hadithHrefs) {
        await page.goto(`https://thaqalayn.net${hadithHref}`);
        await page.waitForLoadState("networkidle");

        const hadithNumMatch = hadithHref.match(/\/(\d+)$/);
        const hadithNumber = hadithNumMatch ? parseInt(hadithNumMatch[1]) : 0;

        // Arabic text — look for RTL / Arabic-script content
        const arabic = await page.evaluate(() => {
            const candidates = document.querySelectorAll("p, div, span");
            for (const el of candidates) {
                const text = el.textContent?.trim() ?? "";
                if (/[؀-ۿ]{10,}/.test(text) && el.children.length === 0)
                    return text;
            }
            return null;
        });

        // English translation — first long LTR paragraph after Arabic
        const english = await page.evaluate(() => {
            const candidates = document.querySelectorAll("p, div");
            let seenArabic = false;
            for (const el of candidates) {
                const text = el.textContent?.trim() ?? "";
                if (/[؀-ۿ]{10,}/.test(text)) { seenArabic = true; continue; }
                if (seenArabic && text.length > 40 && el.children.length === 0) return text;
            }
            return null;
        });

        allHadiths.push({
            book_name: bookNameActual,
            book_number: bookNumber,
            chapter_number: chapterNumber,
            chapter_name: chapterName,
            hadith_number: hadithNumber,
            arabic: arabic ?? null,
            english: english ?? null,
            url: `https://thaqalayn.net${hadithHref}`,
        });

        console.log(`    Hadith ${hadithNumber} ✓`);
    }
}

// 7. Save
fs.writeFileSync(outputPath, JSON.stringify(allHadiths, null, 2), "utf-8");
console.log(`\nDone. ${allHadiths.length} ahadith saved to ${outputPath}`);

await browser.close();
