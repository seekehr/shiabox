import * as fs from "fs";
import * as path from "path";
import { findBookHref, findVolumeHrefs } from "./browser.ts";
import { parseChapterLinks, parseChapterMeta, parseHadithLinks, parseHadithPage } from "./parser.ts";
import { prompt, normalize, slugify, fetchHtml, pool } from "./utils.ts";
import type { Hadith, HadithTask } from "./types.ts";

const CONCURRENCY = 12;

// 1. Prompt and find book
const bookQuery = normalize(await prompt("Enter book name: "));
const book = await findBookHref(bookQuery);

if (!book) {
    console.error(`No book found matching "${bookQuery}".`);
    process.exit(1);
}
console.log(`Found: "${book.name}" → ${book.href}`);

// 2. Create output file
const outputDir = path.join(import.meta.dirname, "..", "output");
fs.mkdirSync(outputDir, { recursive: true });
const outputPath = path.join(outputDir, `${slugify(book.name)}.json`);
fs.writeFileSync(outputPath, "[]", "utf-8");
console.log(`Output: ${outputPath}`);

// 3. Determine volumes to scrape
console.log("Checking for volumes…");
const volumeHrefs = await findVolumeHrefs(book.href);
console.log(`Volumes: ${volumeHrefs.length}`);

const firstBookHtml = await fetchHtml(`https://thaqalayn.net${volumeHrefs[0]}`);
if (!firstBookHtml) { console.error("Failed to fetch book page."); process.exit(1); }

// 4–5. For each volume: fetch book page → chapters → hadith tasks
const hadithTasks: HadithTask[] = [];

for (let vi = 0; vi < volumeHrefs.length; vi++) {
    const volumeHref = volumeHrefs[vi];
    const volumeLabel = volumeHrefs.length > 1 ? ` (volume ${vi + 1}/${volumeHrefs.length})` : "";

    const bookHtml = vi === 0 ? firstBookHtml : (await fetchHtml(`https://thaqalayn.net${volumeHref}`));
    if (!bookHtml) { console.error(`Failed to fetch book page for volume ${vi + 1}.`); continue; }

    const chapterHrefs = parseChapterLinks(bookHtml);
    console.log(`Chapters found${volumeLabel}: ${chapterHrefs.length}`);

    const bookNumMatch = chapterHrefs[0]?.match(/\/chapter\/\d+\/(\d+)\//);
    const bookNumber = bookNumMatch ? parseInt(bookNumMatch[1]) : 0;

    console.log(`Fetching chapter pages${volumeLabel}…`);
    const chapterHtmls = await pool(
        chapterHrefs.map((href) => () => fetchHtml(`https://thaqalayn.net${href}`).then((h) => h ?? "")),
        CONCURRENCY
    );

    for (let ci = 0; ci < chapterHrefs.length; ci++) {
        const chapterHref = chapterHrefs[ci];
        const html = chapterHtmls[ci];
        const chapterNumMatch = chapterHref.match(/\/chapter\/\d+\/\d+\/(\d+)/);
        const chapterNumber = chapterNumMatch ? parseInt(chapterNumMatch[1]) : ci + 1;
        const { chapterName, hadithCount } = parseChapterMeta(html);
        const hadithHrefs = parseHadithLinks(html);

        const volumeNumber = vi + 1;
        if (hadithHrefs.length > 0) {
            for (const href of hadithHrefs) {
                const numMatch = href.match(/\/(\d+)$/);
                hadithTasks.push({ href, bookName: book.name, bookNumber, volumeNumber, chapterNumber, chapterName, hadithNumber: numMatch ? parseInt(numMatch[1]) : 0 });
            }
        } else {
            const [, bookId, bNum, cNum] = chapterHref.split("/").filter(Boolean);
            for (let h = 1; h <= (hadithCount || 1); h++) {
                hadithTasks.push({ href: `/hadith/${bookId}/${bNum}/${cNum}/${h}`, bookName: book.name, bookNumber, volumeNumber, chapterNumber, chapterName, hadithNumber: h });
            }
        }
    }
}

console.log(`Total ahadith to fetch: ${hadithTasks.length}`);

// 6. Fetch and parse all hadith pages in parallel
let done = 0;
const hadiths = await pool(
    hadithTasks.map((task) => async () => {
        const url = `https://thaqalayn.net${task.href}`;
        const html = await fetchHtml(url);
        done++;
        if (done % 50 === 0 || done === hadithTasks.length)
            process.stdout.write(`\r  ${done}/${hadithTasks.length} ahadith fetched`);
        if (!html) return null;
        return parseHadithPage(html, { bookName: task.bookName, bookNumber: task.bookNumber, volumeNumber: task.volumeNumber, chapterNumber: task.chapterNumber, chapterName: task.chapterName, hadithNumber: task.hadithNumber, url });
    }),
    CONCURRENCY
);

// 7. Sort and save
const validHadiths = (hadiths.filter((h) => h !== null) as Hadith[])
    .sort((a, b) => a.chapter_number - b.chapter_number || a.hadith_number - b.hadith_number);

fs.writeFileSync(outputPath, JSON.stringify(validHadiths, null, 2), "utf-8");
console.log(`\n\nDone. ${validHadiths.length} ahadith saved to ${outputPath}`);
