import { chromium } from "playwright";
import { normalize } from "./utils.ts";

export async function findVolumeHrefs(bookHref: string): Promise<string[]> {
    const browser = await chromium.launch({ headless: true });
    const page = await browser.newPage();

    await page.goto(`https://thaqalayn.net${bookHref}`);
    await page.waitForLoadState("networkidle");

    const combobox = page.locator('[role="combobox"]').filter({ hasText: /Volume\s+\d+/i });
    if (await combobox.count() === 0) {
        await browser.close();
        return [bookHref];
    }

    // Count options without navigating
    await combobox.click();
    await page.waitForSelector('[role="option"]');
    const count = await page.locator('[role="option"]').count();
    await page.keyboard.press("Escape");

    const hrefs: string[] = [];

    for (let i = 0; i < count; i++) {
        await combobox.click();
        await page.waitForSelector('[role="option"]');
        await page.locator('[role="option"]').nth(i).click();
        await page.waitForLoadState("networkidle");
        hrefs.push(new URL(page.url()).pathname);
    }

    await browser.close();
    return hrefs;
}

export async function findBookHref(query: string): Promise<{ href: string; name: string } | null> {
    const browser = await chromium.launch({ headless: true });
    const page = await browser.newPage();

    await page.goto("https://thaqalayn.net");
    await page.waitForLoadState("networkidle");

    const bookLinks = page.locator("a[href^='/book/']");
    const count = await bookLinks.count();

    for (let i = 0; i < count; i++) {
        const link = bookLinks.nth(i);
        const titleEl = link.locator("div.font-bold");
        if (await titleEl.count() === 0) continue;
        const raw = (await titleEl.textContent()) ?? "";
        if (normalize(raw).includes(query)) {
            const href = await link.getAttribute("href");
            await browser.close();
            if (href) return { href, name: raw.trim() };
        }
    }

    await browser.close();
    return null;
}
