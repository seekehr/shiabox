import React, { useState, useRef, useEffect } from 'react';
import { Search, User, X } from 'lucide-react';
import Navbar from '../components/Navbar';
import { sendAIPrompt } from '../controller/ai_controller';
import ParseMD from '../utils/ParseAIOutput';
import zulifqarIcon from '../assets/zulifqar.jpg';

const searchSuggestions = [
	"Why do Shias say \"Ya Ali\"?",
	"What are some signs of the Mahdi?",
	"What is the significance of knowledge?",
	"When was 'Ali ibn Musa (AS) born?",
	"I'm so lost. What do I do?",
	"Racism in ahadith."
]

interface Message {
	sender: 'user' | 'ai';
	text: string;
	isError?: boolean;
}

const SearchPage = () => {
	const [searchQuery, setSearchQuery] = useState('');
	const [messages, setMessages] = useState<Message[]>([]);
	const [isSearching, setIsSearching] = useState(false);
	const [abortController, setAbortController] = useState<AbortController | null>(null);
	const messagesEndRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
	}, [messages]);

	const handleStop = () => {
		if (abortController) {
			abortController.abort();
			setIsSearching(false);
			setAbortController(null);
		}
	};

	const handleSearch = (e: React.FormEvent) => {
		e.preventDefault();
		const currentQuery = searchQuery.trim();
		if (!currentQuery) return;

		setIsSearching(true);
		setMessages(prev => [...prev, { sender: 'user', text: currentQuery }, { sender: 'ai', text: '' }]);
		setSearchQuery('');

		const controller = sendAIPrompt(currentQuery,
			(streamData) => {
				if (streamData.done) {
					setIsSearching(false);
					setAbortController(null);
					return;
				}
				setMessages(prev => {
					const lastMessage = prev[prev.length - 1];
					if (lastMessage.sender === 'ai') {
						const newMessages = [...prev];
						const content = streamData.data?.choices?.[0]?.delta?.content || '';
						newMessages[newMessages.length - 1] = {
							...lastMessage,
							text: lastMessage.text + content,
						};
						return newMessages;
					}
					return prev;
				});
			},
			(error) => {
				console.error(error);
				setMessages(prev => {
					const newMessages = [...prev];
					const lastMessage = newMessages[newMessages.length - 1];
					if (lastMessage.sender === 'ai') {
						newMessages[newMessages.length - 1] = {
							...lastMessage,
							text: error,
							isError: true,
						};
						return newMessages;
					}
					return [...newMessages, { sender: 'ai', text: error, isError: true }];
				});
				setIsSearching(false);
				setAbortController(null);
			}
		);
		setAbortController(controller);
	};

	const hasMessages = messages.length > 0;

	return (
		<div className="flex flex-col min-h-screen bg-background">
			<Navbar />

			<main className={`flex-grow flex flex-col px-6 md:px-12 lg:px-24 pt-24 pb-12 transition-all duration-500 ${
				hasMessages ? 'justify-start' : 'justify-center'
			}`}>
				{/* Asymmetric layout: wide left margin */}
				<div className="w-full max-w-3xl ml-0 md:ml-[8vw] lg:ml-[12vw]">

					{/* Hero / Title */}
					{!hasMessages && (
						<div className="mb-16" style={{ animation: 'fade-in-up 0.6s ease-out' }}>
							<p className="font-[Inter] text-[11px] font-medium uppercase tracking-[0.08em] text-on-surface-variant/70 mb-4">
								AI-Powered Hadith Search
							</p>
							<h1 className="font-[Manrope] text-[3.5rem] md:text-[4.5rem] font-light leading-[1.05] tracking-[0.02em] text-on-surface mb-6">
								Shiabox
							</h1>
							<p className="font-[Inter] text-base font-light leading-relaxed text-on-surface-variant max-w-md">
								Search for any hadith or topic across authenticated Shia collections. Still a work in progress.
							</p>
						</div>
					)}

					{/* Search input */}
					<form onSubmit={handleSearch} className={`relative ${hasMessages ? 'mb-10' : 'mb-14'}`}>
						<div className="relative group">
							<input
								type="text"
								value={searchQuery}
								onChange={(e) => setSearchQuery(e.target.value)}
								placeholder="Ask about any hadith or topic..."
								className="w-full bg-transparent border-0 border-b border-outline-variant/30 px-1 py-4 pr-14 text-base font-[Inter] font-light text-on-surface placeholder-on-surface-variant/40 focus:border-primary focus:outline-none transition-all duration-300"
								disabled={isSearching}
							/>
							<button
								type={isSearching ? "button" : "submit"}
								onClick={isSearching ? handleStop : undefined}
								disabled={isSearching ? false : !searchQuery.trim()}
								className="absolute right-0 top-1/2 -translate-y-1/2 p-2.5 rounded-[0.375rem] bg-gradient-to-br from-primary to-primary-container text-on-primary-fixed disabled:opacity-30 disabled:cursor-not-allowed hover:opacity-90 transition-all duration-300"
							>
								{isSearching ? (
									<X className="w-4 h-4" />
								) : (
									<Search className="w-4 h-4" />
								)}
							</button>
						</div>
					</form>

					{/* Messages or Suggestions */}
					{hasMessages ? (
						<div className="space-y-8">
							{messages.map((msg, index) => (
								<div
									key={index}
									className="flex items-start gap-4"
									style={{ animation: 'fade-in-up 0.4s ease-out' }}
								>
									{msg.sender === 'ai' ? (
										<img
											src={zulifqarIcon}
											alt="Shiabox AI"
											className="w-7 h-7 rounded-[0.375rem] object-cover flex-shrink-0 mt-0.5"
										/>
									) : (
										<div className="w-7 h-7 rounded-[0.375rem] bg-surface-container flex items-center justify-center flex-shrink-0 mt-0.5">
											<User className="w-3.5 h-3.5 text-on-surface-variant" />
										</div>
									)}

									<div className="flex-1 min-w-0">
										<span className="font-[Inter] text-[10px] font-medium uppercase tracking-[0.06em] text-on-surface-variant/50 mb-2 block">
											{msg.sender === 'ai' ? 'Shiabox' : 'You'}
										</span>
										{msg.sender === 'ai' ? (
											msg.isError ? (
												<p className="font-[Inter] text-sm font-light text-error">{msg.text}</p>
											) : msg.text ? (
												<ParseMD content={msg.text} />
											) : (
												<div className="flex items-center gap-1.5 py-2">
													<div className="w-1.5 h-1.5 rounded-full bg-primary" style={{ animation: 'pulse-dot 1.4s ease-in-out infinite' }} />
													<div className="w-1.5 h-1.5 rounded-full bg-primary" style={{ animation: 'pulse-dot 1.4s ease-in-out 0.2s infinite' }} />
													<div className="w-1.5 h-1.5 rounded-full bg-primary" style={{ animation: 'pulse-dot 1.4s ease-in-out 0.4s infinite' }} />
												</div>
											)
										) : (
											<p className="font-[Inter] text-sm font-light leading-relaxed text-on-surface">{msg.text}</p>
										)}
									</div>
								</div>
							))}
							<div ref={messagesEndRef} />
						</div>
					) : (
						<div className="grid grid-cols-1 md:grid-cols-2 gap-3" style={{ animation: 'fade-in-up 0.6s ease-out 0.15s both' }}>
							{searchSuggestions.map((suggestion, index) => (
								<button
									key={index}
									onClick={() => setSearchQuery(suggestion)}
									className="group relative p-4 rounded-[0.25rem] bg-surface-container-low text-left transition-all duration-300 hover:bg-surface-container hover:scale-[1.01]"
								>
									<div className="flex items-center gap-3">
										<Search className="w-3.5 h-3.5 text-on-surface-variant/40 group-hover:text-primary transition-colors duration-300 flex-shrink-0" />
										<span className="font-[Inter] text-sm font-light text-on-surface-variant group-hover:text-on-surface transition-colors duration-300">
											{suggestion}
										</span>
									</div>
								</button>
							))}
						</div>
					)}

					{/* Disclaimer */}
					{!hasMessages && (
						<div className="mt-20">
							<p className="font-[Inter] text-[10px] font-medium uppercase tracking-[0.06em] text-on-surface-variant/30">
								Note: Results may contain inaccuracies. Verify with authenticated sources.
							</p>
						</div>
					)}
				</div>
			</main>
		</div>
	);
};

export default SearchPage;
