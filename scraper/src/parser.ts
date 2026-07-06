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
    const chapterName = root.querySelector("h1")?.textContent?.trim() ?? "";
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

const ARABIC_RE = /[؀-ۿ]/;

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
    root.querySelectorAll("nav, footer, header, script, style").forEach((el) => el.remove());

    const paragraphs = root
        .querySelectorAll("p, div, span, li")
        .filter((el) => el.childNodes.every((n) => n.nodeType === 3 || (n as any).tagName === undefined || (n as any).childNodes?.length === 0))
        .map((el) => el.textContent?.trim() ?? "")
        .filter((t) => t.length > 20);

    let arabic: string | null = null;
    let english: string | null = null;

    for (const t of paragraphs) {
        if (!arabic && ARABIC_RE.test(t)) { arabic = t; continue; }
        if (arabic && !english && !ARABIC_RE.test(t) && t.length > 30) { english = t; break; }
    }

    return {
        book_name: meta.bookName,
        book_number: meta.bookNumber,
        chapter_number: meta.chapterNumber,
        chapter_name: meta.chapterName,
        hadith_number: meta.hadithNumber,
        arabic,
        english,
        url: meta.url,
    };
}
