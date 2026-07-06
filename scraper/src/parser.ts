import { parse as parseHtml } from "node-html-parser";
import type { Hadith } from "./types.ts";

export function parseChapterLinks(html: string): string[] {
    const root = parseHtml(html);
    return root
        .querySelectorAll("a[href]")
        .map((a) => a.getAttribute("href") ?? "")
        .filter((h) => h.startsWith("/chapter/"))
        .filter((h, i, arr) => arr.indexOf(h) === i);
}

export function parseChapterMeta(html: string): { chapterName: string; hadithCount: number } {
    const root = parseHtml(html);
    const chapterName = root.querySelector("div.text-2xl.mt-2")?.textContent?.trim() ?? "";
    const countMatch = root.rawText.match(/(\d+)\s*[AĀā]?[hḥ]ad[iī]th/i);
    const hadithCount = countMatch ? parseInt(countMatch[1]) : 0;
    return { chapterName, hadithCount };
}

export function parseHadithLinks(html: string): string[] {
    const root = parseHtml(html);
    return root
        .querySelectorAll("a[href]")
        .map((a) => a.getAttribute("href") ?? "")
        .filter((h) => h.startsWith("/hadith/"))
        .filter((h, i, arr) => arr.indexOf(h) === i);
}

export function parseHadithPage(
    html: string,
    meta: {
        bookName: string;
        bookNumber: number;
        chapterNumber: number;
        chapterName: string;
        hadithNumber: number;
        url: string;
    }
): Hadith {
    const root = parseHtml(html);

    const arabicEl = root.querySelector("p[dir='rtl']");
    const englishEl = root.querySelector("p.nassim:not([dir='rtl'])");

    const arabic = arabicEl?.textContent?.trim() || null;
    const english = englishEl?.textContent?.trim() || null;

    return {
        book_name: meta.bookName,
        chapter_number: meta.chapterNumber,
        hadith_number: meta.hadithNumber,
        content: english,
        metadata: {
            book_number: meta.bookNumber,
            chapter_name: meta.chapterName,
            arabic,
            url: meta.url,
        },
    };
}
