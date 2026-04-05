import { Github, BookOpen } from "lucide-react";

const githubUrl = "https://github.com/seekehr/shiabox";

function Navbar() {
    return (
        <nav className="fixed top-0 left-0 right-0 z-50 flex items-center justify-between px-8 py-4 bg-surface/70 backdrop-blur-[20px]">
            <a href="/" className="flex items-center gap-3 group">
                <BookOpen className="w-5 h-5 text-primary transition-colors duration-300" />
                <span className="font-[Manrope] text-lg font-light tracking-[0.05em] text-on-surface transition-colors duration-300 group-hover:text-primary">
                    Shiabox
                </span>
            </a>
            <div className="flex items-center gap-1">
                <a
                    href={githubUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-2.5 rounded-[0.375rem] text-on-surface-variant hover:text-on-surface hover:bg-surface-container-high/50 transition-all duration-300"
                >
                    <Github className="w-[18px] h-[18px]" />
                </a>
            </div>
        </nav>
    );
}

export default Navbar;
