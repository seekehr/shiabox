export interface Hadith {
    book_name: string;
    chapter_number: number;
    hadith_number: number;
    content: string | null;
    metadata: {
        book_number: number;
        chapter_name: string;
        arabic: string | null;
        url: string;
    };
}

export interface HadithTask {
    href: string;
    bookName: string;
    bookNumber: number;
    chapterNumber: number;
    chapterName: string;
    hadithNumber: number;
}
