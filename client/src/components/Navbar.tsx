import { Home, Sparkles, Github } from "lucide-react";

const githubUrl = "https://github.com/seekehr/shiabox";

function Navbar() {
    return (
        <nav className="flex items-center justify-between px-6 py-4 bg-gray-900/80 backdrop-blur-md border-b border-gray-700/50">
            <div className="flex items-center space-x-3">
                <h1 className="text-xl font-semibold text-white">Shiabox</h1>
            </div>
            <div className="flex items-center space-x-2">
            <a
                    href={githubUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-2 rounded-full hover:bg-gray-800 transition-colors duration-200"
                >
                    <Github className="w-5 h-5 text-gray-300 hover:text-white" />
                </a>
                <button className="p-2 rounded-full hover:bg-gray-800 transition-colors duration-200">
                    <Home className="w-5 h-5 text-gray-300 hover:text-white" />
                </button>
            </div>
        </nav>
    );
}

export default Navbar;