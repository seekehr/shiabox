import { chromium } from "playwright";
import { normalize } from "./utils.ts";

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
