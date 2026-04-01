import React, { useState } from 'react';
import { Search, User, Bot, X } from 'lucide-react';
import Navbar from '../components/Navbar';
import { sendAIPrompt } from '../controller/ai_controller';
import type { StreamResponse } from '../controller/controllers';
import ParseMD from '../utils/ParseAIOutput';

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
			(streamData) => { // onReceive
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
			(error) => { // onError
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

	return (
		<div className="flex flex-col min-h-screen bg-gradient-to-br from-slate-900 via-gray-900 to-slate-800">
			<Navbar />
			<main className="flex-grow flex flex-col items-center justify-center px-6">
				<div className="w-full max-w-4xl mx-auto">
					{/* Title text */}
					<div className="text-center mb-12">
						<h2 className="text-4xl md:text-5xl font-bold text-white mb-4">
							Shiabox AI Search
						</h2>
						<p className="text-lg text-gray-300 max-w-2xl mx-auto">
							Search for any hadith or topic. WIP (contribute on github)
						</p>
					</div>

					{/* input search form */}
					<form onSubmit={handleSearch} className="relative mb-8">
						<div className="relative">
							<input
								type="text"
								value={searchQuery}
								onChange={(e) => setSearchQuery(e.target.value)}
								placeholder="Ask about any hadith or topic..."
								className="w-full px-6 py-4 pr-16 text-lg rounded-2xl border-2 border-gray-600 bg-gray-800/90 backdrop-blur-sm text-white placeholder-gray-400 focus:border-emerald-500 focus:outline-none transition-all duration-200 shadow-xl hover:shadow-2xl hover:bg-gray-800"
								disabled={isSearching}
							/>
							<button
								type={isSearching ? "button" : "submit"}
								onClick={isSearching ? handleStop : undefined}
								disabled={isSearching ? false : !searchQuery.trim()}
								className="absolute right-2 top-1/2 transform -translate-y-1/2 p-3 rounded-xl bg-gradient-to-r from-emerald-500 to-teal-600 text-white hover:from-emerald-600 hover:to-teal-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-lg hover:shadow-xl"
							>
								{isSearching ? (
									<X className="w-5 h-5" />
								) : (
									<Search className="w-5 h-5" />
								)}
							</button>
						</div>
					</form>

					{/* messages bw AI and user; prob gonna move this to /chat?id route */}
					{messages.length > 0 ? (
						<div className="space-y-6 max-w-4xl w-full">
							{messages.map((msg, index) => (
								<div key={index} className={`flex items-start gap-4 ${msg.sender === 'user' ? 'justify-end' : ''}`}>
									{msg.sender === 'ai' && (
										<div className="w-8 h-8 rounded-full bg-gray-700 flex items-center justify-center flex-shrink-0">
											<Bot className="w-5 h-5 text-emerald-400" />
										</div>
									)}
									<div className={`p-4 rounded-2xl max-w-2xl ${msg.sender === 'user' ? 'bg-emerald-600 text-white' : 'bg-gray-800 text-gray-200'}`}>
										{msg.sender === 'ai' ? (
											msg.isError ? (
												<p className="whitespace-pre-wrap text-red-500">{msg.text}</p>
											) : (
												<ParseMD content={msg.text} />
											)
										) : (
											<p className="whitespace-pre-wrap">{msg.text}</p>
										)}
									</div>
									{msg.sender === 'user' && (
										<div className="w-8 h-8 rounded-full bg-gray-700 flex items-center justify-center flex-shrink-0">
											<User className="w-5 h-5 text-gray-300" />
										</div>
									)}
								</div>
							))}
							{isSearching && (
								<div className="flex items-start gap-4">
									<div className="w-8 h-8 rounded-full bg-gray-700 flex items-center justify-center flex-shrink-0">
										<Bot className="w-5 h-5 text-emerald-400" />
									</div>
									<div className="p-4 rounded-2xl max-w-2xl bg-gray-800 text-gray-200">
										<div className="flex items-center space-x-2">
											<div className="w-2 h-2 bg-emerald-400 rounded-full animate-pulse"></div>
											<div className="w-2 h-2 bg-emerald-400 rounded-full animate-pulse delay-75"></div>
											<div className="w-2 h-2 bg-emerald-400 rounded-full animate-pulse delay-150"></div>
										</div>
									</div>
								</div>
							)}
						</div>
					) : (
						<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-12">
							{searchSuggestions.map((suggestion, index) => (
								<button
									key={index}
									onClick={() => setSearchQuery(suggestion)}
									className="p-4 rounded-xl bg-gray-800/80 backdrop-blur-sm border border-gray-700/50 hover:border-emerald-400 hover:bg-gray-800 transition-all duration-200 text-left group shadow-lg hover:shadow-xl"
								>
									<div className="flex items-center space-x-3">
										<div className="p-2 rounded-lg bg-gradient-to-r from-emerald-900/50 to-teal-900/50 group-hover:from-emerald-800/60 group-hover:to-teal-800/60 transition-colors duration-200">
											<Search className="w-4 h-4 text-emerald-400" />
										</div>
										<span className="text-gray-200 font-medium group-hover:text-white">{suggestion}</span>
									</div>
								</button>
							))}
						</div>
					)}

					{/* disclaimer */}
					<div className="text-center text-gray-400 text-sm mt-12">
						<p>NOTE: TAKE AS A GRAIN OF SALT. MAY HALLUCINATE AND INAUTHENTIC AHADITH ARE ALSO INCLUDED.</p>
					</div>
				</div>
			</main>
		</div>
	);
};

export default SearchPage;
