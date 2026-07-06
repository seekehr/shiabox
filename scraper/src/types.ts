export interface Hadith {
    book_name: string;
    book_number: number;
    chapter_number: number;
    chapter_name: string;
    hadith_number: number;
    arabic: string | null;
    english: string | null;
    url: string;
}

export interface HadithTask {
    href: string;
    bookName: string;
    bookNumber: number;
    chapterNumber: number;
    chapterName: string;
    hadithNumber: number;
}
