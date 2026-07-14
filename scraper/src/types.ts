export interface Hadith {
    book_name: string;
    volume_number: number;
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
    volumeNumber: number;
    chapterNumber: number;
    chapterName: string;
    hadithNumber: number;
}
